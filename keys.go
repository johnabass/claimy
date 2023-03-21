package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type KeyHandler struct {
	l *zap.Logger
	k jwk.Key
	s jwk.Set
}

func (kh KeyHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var (
		body   []byte
		err    error
		accept = request.Header.Get("Accept")
	)

	switch accept {
	case "application/x-pem-file":
		body, err = jwk.EncodePEM(kh.k)

	case "application/jwk-set+json":
		body, err = json.Marshal(kh.s)

	case "":
		fallthrough
	case "*/*":
		fallthrough
	case "application/jwk+json":
		body, err = json.Marshal(kh.k)

	default:
		kh.l.Error("unsupported media type requested by client", zap.String("Accept", accept))
		response.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if err == nil {
		_, err = response.Write(body)
	}

	if err != nil {
		kh.l.Error("unable to write key to response", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func provideKey() fx.Option {
	return fx.Options(
		fx.Provide(
			func() (key jwk.Key, err error) {
				var (
					pk  *ecdsa.PrivateKey
					kid [16]byte
				)

				pk, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
				if err == nil {
					key, err = jwk.FromRaw(pk)
				}

				if err == nil {
					_, err = io.ReadFull(rand.Reader, kid[:])
				}

				err = key.Set(jwk.KeyIDKey, base64.RawURLEncoding.EncodeToString(kid[:]))
				return
			},
			func(l *zap.Logger, key jwk.Key) (kh KeyHandler, err error) {
				var pub jwk.Key
				pub, err = key.PublicKey()

				if err == nil {
					kh.l = l
					kh.k = pub
					kh.s = jwk.NewSet()
					kh.s.AddKey(pub)
				}

				return
			},
		),
	)
}
