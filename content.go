package main

import (
	"embed"
	"net/http"

	"go.uber.org/fx"
)

//go:embed content
var content embed.FS

type ContentHandler http.Handler

func provideContent() fx.Option {
	return fx.Options(
		fx.Supply(content),
		fx.Provide(
			func(content embed.FS) http.FileSystem {
				return http.FS(content)
			},
			func(root http.FileSystem) ContentHandler {
				return http.FileServer(root)
			},
		),
	)
}
