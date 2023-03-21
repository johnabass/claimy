package main

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Configuration struct {
	Address string
	Scripts []string
}

func provideConfiguration() fx.Option {
	return fx.Options(
		fx.Provide(
			func() (v *viper.Viper, err error) {
				v = viper.New()
				v.SetConfigName("claimy")
				v.AddConfigPath(".")
				v.AddConfigPath("$HOME/.claimy")
				v.AddConfigPath("/etc/claimy")

				err = v.ReadInConfig()
				return
			},
			func(v *viper.Viper) (cfg Configuration, err error) {
				err = v.Unmarshal(&cfg)
				return
			},
		),
	)
}
