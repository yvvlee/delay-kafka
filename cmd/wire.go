//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package cmd

import (
	"github.com/google/wire"

	"github.com/yvvlee/delay-kafka/internal/config"
	"github.com/yvvlee/delay-kafka/internal/core"
)

func wireApp() (*core.App, func(), error) {
	panic(
		wire.Build(
			config.NewConfig,
			core.ProviderSet,
		),
	)
}
