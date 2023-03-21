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

	KeyHandler     KeyHandler
	KeySetHandler  KeySetHandler
	IssueHandler   IssueHandler
	ContentHandler ContentHandler
}

func provideServer() fx.Option {
	return fx.Options(
		fx.Provide(
			func(in serveMuxIn) (mux *http.ServeMux) {
				mux = http.NewServeMux()

				mux.Handle("/keys/"+in.KeyHandler.key.KeyID(), in.KeyHandler)
				mux.Handle("/keys", in.KeySetHandler)
				mux.Handle("/issue", in.IssueHandler)
				mux.Handle("/content/", in.ContentHandler)

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
							defer s.Shutdown()
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
