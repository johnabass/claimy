package main

import "go.uber.org/fx"

type Claimer interface {
	Claims()
}

func provideClaimer() fx.Option {
	return fx.Options(
		fx.Provide(
			func(cfg Configuration) (Claimer, error) {
				return nil, nil
			},
		),
	)
}
