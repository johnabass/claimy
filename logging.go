package main

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func provideLogging() fx.Option {
	return fx.Options(
		fx.Provide(
			func() (l *zap.Logger, err error) {
				l, err = zap.NewDevelopment()
				if err == nil {
					_, err = zap.RedirectStdLogAt(l, zapcore.ErrorLevel)
				}

				return
			},
		),
		fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: l}
		}),
	)
}
