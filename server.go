package main

import (
	"context"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func provideServer() fx.Option {
	return fx.Options(
		fx.Provide(
			func(kh KeyHandler) (mux *http.ServeMux) {
				mux = http.NewServeMux()
				mux.Handle("/key", kh)
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
