package server

import (
	apiV1 "azure-vm-backend/api/v1"
	"azure-vm-backend/docs"
	"azure-vm-backend/internal/handler"
	"azure-vm-backend/internal/middleware"
	"azure-vm-backend/pkg/jwt"
	"azure-vm-backend/pkg/log"
	"azure-vm-backend/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHTTPServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	userHandler *handler.UserHandler,
	accountsHandler *handler.AccountsHandler,
	subHandler *handler.SubscriptionsHandler,
	vmHandler *handler.VirtualMachineHandler,
	vmRegionHandler *handler.VmRegionHandler,
	vmImageHandler *handler.VmImageHandler,
) *http.Server {
	gin.SetMode(gin.DebugMode)
	s := http.NewServer(
		gin.Default(),
		logger,
		http.WithServerHost(conf.GetString("http.host")),
		http.WithServerPort(conf.GetInt("http.port")),
	)

	// swagger doc
	docs.SwaggerInfo.BasePath = "/v1"
	s.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		//ginSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", conf.GetInt("app.http.port"))),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	))

	s.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		//middleware.SignMiddleware(log),
	)
	s.GET("/", func(ctx *gin.Context) {
		logger.WithContext(ctx).Info("hello")
		apiV1.HandleSuccess(ctx, map[string]interface{}{
			":)": "Thank you for using Azure-VM-Backend",
		})
	})

	v1 := s.Group("/v1")
	{
		// No route group has permission
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/register", userHandler.Register)
			noAuthRouter.POST("/login", userHandler.Login)
		}
		// Non-strict permission routing group
		noStrictAuthRouter := v1.Group("/").Use(middleware.NoStrictAuth(jwt, logger))
		{
			noStrictAuthRouter.GET("/user", userHandler.GetProfile)
		}

		// Strict permission routing group
		strictAuthRouter := v1.Group("/").Use(middleware.StrictAuth(jwt, logger))
		{
			// 用户接口
			strictAuthRouter.POST("/user", userHandler.UpdateProfile)
			// 账户接口
			strictAuthRouter.POST("/accounts/create", accountsHandler.CreateAccounts)
			strictAuthRouter.DELETE("/accounts/delete", accountsHandler.DeleteAccounts)
			strictAuthRouter.POST("/accounts/list", accountsHandler.ListAccounts)

			strictAuthRouter.POST("/accounts/update/:id", accountsHandler.UpdateAccount)
			strictAuthRouter.GET("/accounts/:id", accountsHandler.GetAccount)
			strictAuthRouter.POST("/accounts/sync", accountsHandler.SyncAccounts)

			// 订阅接口
			// 获取指定账号的所有订阅
			strictAuthRouter.POST("/subscriptions/get/:accountId", subHandler.GetSubscriptions)
			strictAuthRouter.POST("/subscriptions/list", subHandler.ListSubscriptions)
			// 获取指定订阅的详细信息
			strictAuthRouter.GET("/subscriptions/:accountId/:subscriptionId", subHandler.GetSubscription)
			// 同步指定账号的订阅信息
			strictAuthRouter.POST("/subscriptions/:accountId/sync", subHandler.SyncSubscriptions)
			// 删除指定账号的所有订阅信息
			strictAuthRouter.DELETE("/subscriptions/:accountId", subHandler.DeleteSubscriptions)

			// 虚拟机接口
			// 查询虚拟机列表(支持过滤、分页等)
			strictAuthRouter.GET("/vms", vmHandler.ListVMs)

			// 获取单个虚拟机详细信息
			strictAuthRouter.GET("/vms/:accountId/instance/:vmId", vmHandler.GetVM)

			// 获取指定账号和订阅下的虚拟机列表
			strictAuthRouter.GET("/vms/:accountId/subscription/:subscriptionId", vmHandler.ListVMsBySubscription)

			// 同步指定账号下的所有虚拟机
			strictAuthRouter.POST("/vms/:accountId/sync", vmHandler.SyncVMs)

			// 同步指定订阅下的虚拟机
			strictAuthRouter.POST("/vms/:accountId/subscription/:subscriptionId/sync", vmHandler.SyncVMsBySubscription)

			// 创建虚拟机（预留）
			strictAuthRouter.POST("/vms/:accountId", vmHandler.CreateVM)

			strictAuthRouter.POST("/vms/:accountId/:id/operate", vmHandler.OperateVM)

			// 更新虚拟机dns标签
			strictAuthRouter.POST("/vms/update/dns/:accountId/:ID", vmHandler.UpdateDNSLabel)

			// 获取区域列表
			strictAuthRouter.GET("/vm/regions", vmRegionHandler.ListVmRegions)

			// 获取单个区域详情
			strictAuthRouter.GET("/vm/regions/:id", vmRegionHandler.GetVmRegion)

			// 镜像接口
			strictAuthRouter.GET("/vm/images", vmImageHandler.ListVmImages)
			// 获取单个镜像详情
			strictAuthRouter.GET("/vm/images/:id", vmImageHandler.GetVmImage)
			// 同步镜像
			strictAuthRouter.POST("/vm/images/sync", vmImageHandler.SyncVmImages)
		}
	}

	return s
}
