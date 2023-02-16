package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("api")

func ConfigRouter(router *gin.Engine, cfg config.Config) {
	apiV1 := router.Group("/api/v1")
	authMiddleware, err := jwtGinMiddleware(cfg.SecretKey)
	if err != nil {
		log.Fatalf("jwt auth middleware: %v", err)
	}

	err = authMiddleware.MiddlewareInit()
	if err != nil {
		log.Fatalf("authMiddleware.MiddlewareInit: %v", err)
	}

	// dashboard
	apiV1.GET("/all_areas", GetAllAreas)
	apiV1.GET("/schedulers", GetSchedulersHandler)
	apiV1.GET("/get_index_info", GetIndexInfoHandler)
	apiV1.GET("/get_device_info", GetDeviceInfoHandler)
	apiV1.GET("/get_map_info", GetMapInfoHandler)
	apiV1.GET("/get_device_info_daily", GetDeviceInfoDailyHandler)
	apiV1.GET("/get_diagnosis_days", GetDeviceDiagnosisDailyHandler)
	apiV1.GET("/get_diagnosis_hours", GetDeviceDiagnosisHourHandler)
	apiV1.GET("/get_cache_list", GetCacheListHandler)
	apiV1.GET("/get_retrieval_list", GetRetrievalListHandler)
	apiV1.GET("/get_validation_list", GetValidationListHandler)
	apiV1.POST("/create_application", CreateApplicationHandler)
	apiV1.GET("/get_applications", GetApplicationsHandler)

	// console
	apiV1.GET("/device_binding", DeviceBindingHandler)
	apiV1.GET("/device_unbinding", DeviceUnBindingHandler)
	apiV1.GET("/device_update", DeviceUpdateHandler)
	apiV1.GET("/get_user_device_profile", GetUserDeviceProfileHandler)

	user := apiV1.Group("/user")
	user.POST("/login", authMiddleware.LoginHandler)
	user.POST("/logout", authMiddleware.LogoutHandler)
	user.Use(authMiddleware.MiddlewareFunc())
	user.GET("/refresh_token", authMiddleware.RefreshHandler)
	user.POST("/info", GetUserInfoHandler)

	// admin
	admin := apiV1.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc())
	admin.GET("/cache_list", GetCacheTaskListHandler)
	admin.GET("/cache_info", GetCacheTaskInfoHandler)
	admin.POST("/add_cache", AddCacheTaskHandler)
	admin.POST("/delete_cache", DeleteCacheTaskHandler)
	admin.POST("/delete_device_cache", DeleteCacheTaskByDeviceHandler)
	admin.GET("/get_cache_info", GetCarFileInfoHandler)
	admin.GET("/get_login_log", GetLoginLogHandler)
	admin.GET("/get_operation_log", GetOperationLogHandler)
}
