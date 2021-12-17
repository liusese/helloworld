// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"example.com/helloworld/app/hw/service/internal/biz"
	"example.com/helloworld/app/hw/service/internal/conf"
	"example.com/helloworld/app/hw/service/internal/data"
	"example.com/helloworld/app/hw/service/internal/server"
	"example.com/helloworld/app/hw/service/internal/service"
)

// initApp init kratos application.
func initApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
