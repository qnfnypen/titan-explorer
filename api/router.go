package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("api")

func ConfigRouter(router *gin.Engine, cfg config.Config) {
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
	apiV0.GET("/all_areas", GetAllAreas)
	apiV0.GET("/schedulers", GetSchedulersHandler)
	apiV0.GET("/get_user_device_info", GetUserDeviceInfoHandler)
	apiV0.GET("/get_index_info", GetIndexInfoHandler)
	apiV0.GET("/get_device_info", GetDeviceInfoHandler)
	apiV0.GET("/get_device_info_daily", GetDeviceInfoDailyHandler)
	apiV0.GET("/get_diagnosis_days", GetDeviceDiagnosisDailyHandler)
	apiV0.GET("/get_diagnosis_hours", GetDeviceDiagnosisHourHandler)
	apiV0.GET("/get_cache_list", GetCacheListHandler)
	apiV0.GET("/get_retrieve_list", GetRetrieveListHandler)
	apiV0.POST("/create_application", CreateApplicationHandler)
	apiV0.GET("/get_applications", GetApplicationsHandler)

	// console
	apiV0.GET("/device_binding", DeviceBindingHandler)
	apiV0.GET("/create_task", CreateTaskHandler)
	apiV0.GET("/get_task", GetTaskInfoHandler)
	apiV0.GET("/get_task_list", GetTaskListHandler)
	apiV0.GET("/get_task_detail", GetTaskDetailHandler)

	// admin
	admin := apiV0.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc())
	admin.GET("/cache_task_list", GetCacheTaskListHandler)
	admin.GET("/cache_task_info", GetCacheTaskInfoHandler)
	admin.POST("/add_cache_task", AddCacheTaskHandler)
	admin.POST("/cancel_cache_task", CancelCacheTaskHandler)
	admin.GET("/get_cache_info", GetCarFileInfoHandler)
	admin.POST("/remove_cache", RemoveCacheHandler)
	admin.GET("/get_login_log", GetLoginLogHandler)
	admin.GET("/get_operation_log", GetOperationLogHandler)
}
