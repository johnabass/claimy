package main

import (
	"embed"
	"net/http"

	"go.uber.org/fx"
)

//go:embed swaggerui
var swaggerui embed.FS

//go:embed swagger.yaml
var swaggerYAML []byte

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
			func() SwaggerYAMLHandler {
				return SwaggerYAMLHandler(
					http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
						//response.Header().Set("Content-Type", "application/yaml")
						response.Write(swaggerYAML)
					}),
				)
			},
		),
	)
}
