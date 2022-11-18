package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("api")

func ServerAPI(cfg *config.Config) {
	gin.SetMode(cfg.Mode)

	if cfg.Mode == gin.DebugMode {
		logging.SetDebugLogging()
	}

	router := gin.Default()
	router.Use(cors.Default())

	apiV0 := router.Group("/api/v0")
	authMiddleware, err := jwtGinMiddleware(cfg.SecretKey)
	if err != nil {
		log.Fatalf("jwt auth middleware: %v", err)
	}

	err = authMiddleware.MiddlewareInit()
	if err != nil {
		log.Fatalf("authMiddleware.MiddlewareInit: %v", err)
	}

	apiV0.POST("/login", authMiddleware.LoginHandler)
	apiV0.POST("/logout", authMiddleware.LogoutHandler)
	apiV0.GET("/refresh_token", authMiddleware.RefreshHandler)

	// dashboard
	apiV0.GET("/schedulers", GetSchedulersHandler)
	apiV0.GET("/get_miner_info", GetAllMinerInfoHandler)
	apiV0.GET("/get_retrieval", RetrievalHandler)
	apiV0.GET("/get_user_device_info", GetUserDeviceInfoHandler)
	apiV0.GET("/get_index_info", GetIndexInfoHandler)
	apiV0.GET("/get_device_info", GetDeviceInfoHandler)
	apiV0.GET("/get_diagnosis_days", GetDeviceDiagnosisDailyHandler)
	apiV0.GET("/get_diagnosis_hours", GetDeviceDiagnosisHourHandler)

	// console
	{
		apiV0.GET("/device_biding", DeviceBidingHandler)
		apiV0.GET("/device_create", DeviceCreateHandler)
		apiV0.GET("/create_task", CreateTaskHandler)
		apiV0.GET("/get_task", GetTaskInfoHandler)
		apiV0.GET("/get_task_list", GetTaskListHandler)
		apiV0.GET("/get_task_detail", GetTaskDetailHandler)
	}

	// admin
	admin := apiV0.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc())
	{
		admin.POST("/add_scheduler", nil)
		admin.POST("/delete_scheduler", nil)
	}

	if err := router.Run(cfg.ApiListen); err != nil {
		log.Fatalf("starting server: %v\n", err)
	}
}
