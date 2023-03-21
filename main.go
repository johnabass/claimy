package main

import (
	"fmt"
	"os"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		provideLogging(),
		provideConfiguration(),
		provideKey(),
		provideTokener(),
		provideServer(),
	)

	app.Run()
	if app.Err() != nil {
		fmt.Fprintln(os.Stderr, app.Err())
		os.Exit(1)
	}
}
