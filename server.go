package main

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type serveMuxIn struct {
	fx.In

	KeyHandler         KeyHandler
	KeySetHandler      KeySetHandler
	IssueHandler       IssueHandler
	SwaggerUIHandler   SwaggerUIHandler
	SwaggerYAMLHandler SwaggerYAMLHandler
}

func provideServer() fx.Option {
	return fx.Options(
		fx.Provide(
			func(in serveMuxIn) (mux *http.ServeMux) {
				mux = http.NewServeMux()

				mux.Handle("/keys/"+in.KeyHandler.key.KeyID(), in.KeyHandler)
				mux.Handle("/keys", in.KeySetHandler)
				mux.Handle("/issue", in.IssueHandler)
				mux.Handle("/swaggerui/", in.SwaggerUIHandler)
				mux.Handle("/openapi.yaml", in.SwaggerYAMLHandler)

				return
			},
			func(l *zap.Logger, cfg Configuration, mux *http.ServeMux) *http.Server {
				return &http.Server{
					Addr:    cfg.Address,
					Handler: mux,
				}
			},
		),
		fx.Invoke(
			func(l fx.Lifecycle, s fx.Shutdowner, logger *zap.Logger, server *http.Server) {
				l.Append(fx.Hook{
					OnStart: func(context.Context) error {
						go func() {
							defer func() {
								if err := s.Shutdown(); err != nil {
									logger.Error("error shutting down server", zap.Error(err))
								}
							}()

							err := server.ListenAndServe()
							if err != nil && !errors.Is(err, http.ErrServerClosed) {
								logger.Error("error starting server", zap.Error(err))
							}
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						return server.Shutdown(ctx)
					},
				})
			},
		),
	)
}
