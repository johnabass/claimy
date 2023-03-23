package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/robertkrimen/otto"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Tokener interface {
	NewToken() (jwt.Token, error)
}

type IssueHandler struct {
	l   *zap.Logger
	t   Tokener
	key jwk.Key
}

func (ih IssueHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	token, err := ih.t.NewToken()
	if err != nil {
		ih.l.Error("could not generate token", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := jwt.Sign(
		token,
		jwt.WithKey(
			jwa.ES512,
			ih.key,
		),
	)

	if err != nil {
		ih.l.Error("unable to sign token", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	JWTMediaType.SetContentType(response.Header())
	response.Write(body)
}

type tokener struct {
	vm      *otto.Otto
	scripts []*otto.Script
	initial map[string]interface{}

	claims   map[string]interface{}
	validFor time.Duration
}

func (t tokener) NewToken() (token jwt.Token, err error) {
	b := jwt.NewBuilder()
	for k, v := range t.initial {
		b = b.Claim(k, v)
	}

	vm := t.vm.Copy()
	err = vm.Set("builder", b)

	for i := 0; err == nil && i < len(t.scripts); i++ {
		_, err = vm.Run(t.scripts[i])
	}

	for k, v := range t.claims {
		b = b.Claim(k, v)
	}

	now := time.Now()
	b = b.IssuedAt(now)
	if t.validFor > 0 {
		b = b.Expiration(now.Add(t.validFor))
	}

	token, err = b.Build()
	return
}

func provideTokener() fx.Option {
	return fx.Options(
		fx.Provide(
			func() *otto.Otto {
				return otto.New()
			},
			func(l *zap.Logger, vm *otto.Otto, cfg Configuration) (scripts []*otto.Script, err error) {
				if len(cfg.Scripts) > 0 {
					scripts = make([]*otto.Script, 0, len(cfg.Scripts))
					for _, pattern := range cfg.Scripts {
						pattern = os.ExpandEnv(pattern)
						matches, err := filepath.Glob(pattern)
						if err != nil {
							return nil, err
						} else if len(matches) == 0 {
							l.Info("no scripts match pattern", zap.String("pattern", pattern))
							continue
						}

						for _, scriptFile := range matches {
							l.Info("parsing script", zap.String("file", scriptFile))

							f, err := os.Open(scriptFile)
							if err != nil {
								return nil, err
							}

							defer f.Close()
							script, err := vm.Compile(scriptFile, f)
							if err != nil {
								return nil, err
							}

							scripts = append(scripts, script)
						}
					}
				}

				return
			},
			func(l *zap.Logger, vm *otto.Otto, scripts []*otto.Script, cfg Configuration) Tokener {
				l.Info("configured claims",
					zap.Any("claims", cfg.Claims), zap.Duration("validFor", cfg.ValidFor),
				)

				return tokener{
					vm:       vm,
					claims:   cfg.Claims,
					scripts:  scripts,
					validFor: cfg.ValidFor,
				}
			},
			func(l *zap.Logger, t Tokener, k jwk.Key) (ih IssueHandler, err error) {
				ih = IssueHandler{
					l:   l,
					t:   t,
					key: k,
				}

				return
			},
		),
	)
}
