//go:build wireinject
// +build wireinject

package wire

import (
	"azure-vm-backend/internal/handler"
	"azure-vm-backend/internal/repository"
	"azure-vm-backend/internal/server"
	"azure-vm-backend/internal/service"
	"azure-vm-backend/pkg/app"
	"azure-vm-backend/pkg/jwt"
	"azure-vm-backend/pkg/log"
	"azure-vm-backend/pkg/server/http"
	"azure-vm-backend/pkg/sid"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

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

var handlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewUserHandler,
	handler.NewAccountsHandler,
	handler.NewSubscriptionsHandler,
	handler.NewVirtualMachineHandler,
	handler.NewVmRegionHandler,
	handler.NewVmImageHandler,
	handler.NewVmSizeHandler,
)

var serverSet = wire.NewSet(
	server.NewHTTPServer,
	server.NewJob,
)

// build App
func newApp(
	httpServer *http.Server,
	job *server.Job,
	// task *server.Task,
) *app.App {
	return app.NewApp(
		app.WithServer(httpServer, job),
		app.WithName("server"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		serverSet,
		sid.NewSid,
		jwt.NewJwt,
		newApp,
	))
}
