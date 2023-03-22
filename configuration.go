package main

import (
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Configuration struct {
	Address string
	Scripts []string

	Claims   map[string]interface{}
	ValidFor time.Duration
}

func provideConfiguration() fx.Option {
	return fx.Options(
		fx.Provide(
			func(l *zap.Logger) (v *viper.Viper, err error) {
				v = viper.New()
				v.SetConfigName("claimy")
				v.AddConfigPath(".")
				v.AddConfigPath("$HOME/.claimy")
				v.AddConfigPath("/etc/claimy")

				err = v.ReadInConfig()
				var notFoundErr viper.ConfigFileNotFoundError
				switch {
				case err == nil:
					l.Info("configuration file found", zap.String("file", v.ConfigFileUsed()))

				case errors.As(err, &notFoundErr):
					err = nil // ignore
					l.Info("no configuration file")
				}

				return
			},
			func(v *viper.Viper) (cfg Configuration, err error) {
				err = v.Unmarshal(
					&cfg,
					func(dco *mapstructure.DecoderConfig) {
						dco.ErrorUnused = true
					},
				)

				return
			},
		),
	)
}
