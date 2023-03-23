package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	JWKMediaType    = "application/jwk+json"
	JWKSetMediaType = "application/jwk-set+json"
	PEMMediaType    = "application/x-pem-file"
)

type KeySetHandler struct {
	l   *zap.Logger
	set jwk.Set
}

func (ksh KeySetHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	body, err := json.Marshal(ksh.set)
	if err != nil {
		ksh.l.Error("unable to marshal JWK set", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", JWKSetMediaType)
	response.Write(body)
}

type KeyHandler struct {
	l   *zap.Logger
	key jwk.Key
}

func (kh KeyHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	responseMediaType := JWKMediaType
	if accept := request.Header.Get("Accept"); len(accept) > 0 {
		mediaType, params, err := mime.ParseMediaType(accept)
		switch {
		case err != nil:
			kh.l.Error("invalid Accept header", zap.Error(err))
			response.WriteHeader(http.StatusBadRequest)
			return

		case len(params) > 0:
			kh.l.Error("parameters aren't supported in the Accept header")
			response.WriteHeader(http.StatusUnsupportedMediaType)
			return

		case mediaType == JWKMediaType:
			fallthrough

		case mediaType == PEMMediaType:
			responseMediaType = mediaType

		case mediaType == "*/*":
			fallthrough

		case mediaType == "application/*":
			responseMediaType = JWKMediaType

		default:
			kh.l.Error("unsupported media type", zap.String("mediaType", mediaType))
			response.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
	}

	var (
		body []byte
		err  error
	)

	switch responseMediaType {
	case JWKMediaType:
		body, err = json.Marshal(kh.key)

	default:
		body, err = jwk.EncodePEM(kh.key)
	}

	if err != nil {
		kh.l.Error("unable to encode key", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", responseMediaType)
	response.Write(body)
}

func provideKey() fx.Option {
	return fx.Options(
		fx.Provide(
			func(l *zap.Logger) (key jwk.Key, err error) {
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

				if err == nil {
					kidString := base64.RawURLEncoding.EncodeToString(kid[:])
					l.Info("generated key", zap.String("kid", kidString))
					err = key.Set(jwk.KeyIDKey, kidString)
				}

				return
			},
			func(l *zap.Logger, key jwk.Key) (kh KeyHandler, err error) {
				var pub jwk.Key
				pub, err = key.PublicKey()

				if err == nil {
					kh.l = l
					kh.key = pub
				}

				return
			},
			func(l *zap.Logger, key jwk.Key) (khs KeySetHandler, err error) {
				var pub jwk.Key
				pub, err = key.PublicKey()

				if err == nil {
					khs.l = l
					khs.set = jwk.NewSet()
					err = khs.set.AddKey(pub)
				}

				return
			},
		),
	)
}
