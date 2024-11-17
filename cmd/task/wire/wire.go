//go:build wireinject
// +build wireinject

package wire

import (
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/internal/server"
	"azure-vm-backend/internal/service"
	"azure-vm-backend/pkg/app"
	"azure-vm-backend/pkg/jwt"
	"azure-vm-backend/pkg/log"
	"azure-vm-backend/pkg/sid"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

// 添加基础设施提供者集合

var repositorySet = wire.NewSet(
	repository.NewDB,
	//repository.NewRedis,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewUserRepository,
	repository.NewAccountsRepository,
	repository.NewSubscriptionsRepository,
	repository.NewVirtualMachineRepository,
	repository.NewVmRegionRepository,
	repository.NewVmImageRepository,
	repository.NewVmSizeRepository,
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewUserService,
	service.NewAccountsService,
	service.NewSubscriptionsService,
	service.NewVirtualMachineService,
	service.NewVmRegionService,
	service.NewVmImageService,
	service.NewVmSizeService,
)

var serverSet = wire.NewSet(
	server.NewTask,
)

// build App
func newApp(
	task *server.Task,
) *app.App {
	return app.NewApp(
		app.WithServer(task),
		app.WithName("azure-task"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		serverSet,
		serviceSet,
		repositorySet,
		newApp,
		sid.NewSid,
		jwt.NewJwt,
	))
}
