package main

import (
	"context"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type serveMuxIn struct {
	fx.In

	KeyHandler    KeyHandler
	KeySetHandler KeySetHandler
}

func provideServer() fx.Option {
	return fx.Options(
		fx.Provide(
			func(in serveMuxIn) (mux *http.ServeMux) {
				mux = http.NewServeMux()

				mux.Handle("/keys/"+in.KeyHandler.key.KeyID(), in.KeyHandler)
				mux.Handle("/keys", in.KeySetHandler)

				return
			},
			func(l *zap.Logger, mux *http.ServeMux) *http.Server {
				return &http.Server{
					Addr:    ":8080",
					Handler: mux,
				}
			},
		),
		fx.Invoke(
			func(l fx.Lifecycle, s *http.Server) {
				l.Append(fx.Hook{
					OnStart: func(context.Context) error {
						go func() {
							s.ListenAndServe()
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						return s.Shutdown(ctx)
					},
				})
			},
		),
	)
}
