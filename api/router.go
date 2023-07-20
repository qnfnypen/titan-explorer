package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("api")

func ConfigRouter(router *gin.Engine, cfg config.Config) {
	apiV1 := router.Group("/api/v1")
	apiV2 := router.Group("/link")
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
	apiV1.GET("/get_query_info", GetQueryInfoHandler)
	// index info all nodes info from device info
	apiV1.GET("/get_nodes_info", GetNodesInfoHandler)
	apiV1.GET("/get_device_info", GetDeviceInfoHandler)
	apiV1.GET("/get_device_active_info", GetDeviceActiveInfoHandler)
	apiV1.GET("/get_device_status", GetDeviceStatusHandler)
	apiV1.GET("/get_map_info", GetMapInfoHandler)
	apiV1.GET("/get_device_info_daily", GetDeviceInfoDailyHandler)
	apiV1.GET("/get_diagnosis_days", GetDeviceDiagnosisDailyByDeviceIdHandler)
	// by-user_id or all node count
	apiV1.GET("/get_diagnosis_days_user", GetDeviceDiagnosisDailyByUserIdHandler)
	apiV1.GET("/get_diagnosis_hours", GetDeviceDiagnosisHourHandler)
	apiV1.GET("/get_cache_hours", GetCacheHourHandler)
	apiV1.GET("/get_cache_days", GetCacheDaysHandler)
	apiV1.POST("/create_application", CreateApplicationHandler)
	apiV1.GET("/get_applications", GetApplicationsHandler)
	apiV1.GET("/get_application_amount", GetApplicationAmountHandler)
	// node daily count
	apiV1.GET("/get_nodes_days", GetDiskDaysHandler)
	// console
	apiV1.GET("/device_binding", DeviceBindingHandler)
	apiV1.GET("/device_unbinding", DeviceUnBindingHandler)
	apiV1.GET("/device_update", DeviceUpdateHandler)
	apiV1.GET("/get_user_device_profile", GetUserDeviceProfileHandler)
	apiV1.GET("/get_user_device_count", GetUserDevicesCountHandler)
	// request from titan api
	apiV1.GET("/get_cache_list", GetCacheListHandler)
	apiV1.GET("/get_retrieval_list", GetRetrievalListHandler)
	apiV1.GET("/get_validation_list", GetValidationListHandler)
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
	admin.GET("/get_node_daily_trend", GetNodeDailyTrendHandler)
	// storage
	storage := apiV1.Group("/storage")
	storage.POST("/get_verify_code", GetVerifyCodeHandle)
	storage.POST("/register", UserRegister)
	storage.POST("/password_reset", PasswordRest)
	storage.GET("/login_before", BeforeLogin)
	storage.POST("/login", authMiddleware.LoginHandler)
	storage.POST("/logout", authMiddleware.LogoutHandler)
	apiV2.GET("/", GetShareLinkHandler)
	//storage.GET("/link", GetShareLinkHandler)
	storage.GET("/get_link", ShareLinkHandler)
	storage.GET("/get_map_cid", GetMapByCidHandler)
	storage.GET("/get_asset_detail", GetCarFileCountHandler)
	storage.GET("/get_asset_location", GetLocationHandler)
	storage.GET("/share_asset", ShareAssetsHandler)
	storage.GET("/get_asset_status", GetAssetStatusHandler)
	storage.Use(authMiddleware.MiddlewareFunc())
	storage.GET("/get_locateStorage", GetAllocateStorageHandler)
	storage.GET("/get_Storage_size", GetStorageSizeHandler)
	storage.GET("/create_asset", CreateAssetHandler)
	storage.GET("/delete_asset", DeleteAssetHandler)
	storage.GET("/get_asset_info", GetAssetInfoHandler)
	storage.GET("/get_asset_list", GetAssetListHandler)
	storage.GET("/share_status_set", UpdateShareStatusHandler)
	storage.GET("/create_key", CreateKeyHandler)
	storage.GET("/get_keys", GetKeyListHandler)
	storage.GET("/delete_key", DeleteKeyHandler)
	storage.GET("/get_asset_count", GetAssetCountHandler)
	storage.GET("/get_user_info_hour", GetStorageHourHandler)
	storage.GET("/get_user_info_daily", GetStorageDailyHandler)
	storage.GET("/refresh_token", authMiddleware.RefreshHandler)
}
