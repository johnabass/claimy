package main

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Configuration struct {
	Address string
	Scripts []string
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
				if err == nil {
					l.Info("configuration file found", zap.String("file", v.ConfigFileUsed()))
				}

				return
			},
			func(v *viper.Viper) (cfg Configuration, err error) {
				err = v.Unmarshal(&cfg)
				return
			},
		),
	)
}
