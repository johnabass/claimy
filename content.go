package main

import (
	"bytes"
	"embed"
	"net/http"

	"text/template"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"go.uber.org/fx"
)

//go:embed swaggerui
var swaggerui embed.FS

//go:embed swagger.yaml.gt
var swaggerYAML string

type SwaggerUIHandler http.Handler
type SwaggerYAMLHandler http.Handler

func provideContent() fx.Option {
	return fx.Options(
		fx.Provide(
			func() SwaggerUIHandler {
				return http.FileServer(
					http.FS(swaggerui),
				)
			},
			func(key jwk.Key) (syh SwaggerYAMLHandler, err error) {
				var (
					tmplt *template.Template
					yaml  []byte
				)

				tmplt, err = template.New("swagger.yaml").Parse(swaggerYAML)
				if err == nil {
					var output bytes.Buffer
					err = tmplt.Execute(
						&output,
						map[string]any{
							"defaultKeyID": key.KeyID(),
						},
					)

					yaml = output.Bytes()
				}

				if err == nil {
					syh = SwaggerYAMLHandler(
						http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
							// NOTE: don't set a content type, since that seems to mess up browsers
							// since there is, as yet, no official MIME type for YAML
							response.Write(yaml)
						}),
					)
				}

				return
			},
		),
	)
}
