package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("api")

func ConfigRouter(router *gin.Engine, cfg config.Config) {
	router.Use(Cors())
	apiV1 := router.Group("/api/v1")
	apiV2 := router.Group("/api/v2")
	apiV3 := router.Group("/link")

	RegisterV1RouterAuthByKey(router)

	authMiddleware, err := jwtGinMiddleware(cfg.SecretKey)
	if err != nil {
		log.Fatalf("jwt auth middleware: %v", err)
	}

	err = authMiddleware.MiddlewareInit()
	if err != nil {
		log.Fatalf("authMiddleware.MiddlewareInit: %v", err)
	}

	// dashboard
	apiV2.GET("/get_high", GetHigh)
	apiV2.GET("/all_areas", GetAllAreas)
	apiV2.GET("/schedulers", GetSchedulersHandler)
	apiV2.GET("/get_index_info", GetIndexInfoHandler)
	apiV2.GET("/get_query_info", GetQueryInfoHandler)
	// index info all nodes info from device info
	apiV2.GET("/get_nodes_info", GetNodesInfoHandler)
	apiV2.GET("/get_device_info", GetDeviceInfoHandler)
	apiV2.GET("/get_device_status", GetDeviceStatusHandler)
	apiV2.GET("/get_map_info", GetMapInfoHandler)
	apiV2.GET("/get_device_info_daily", GetDeviceInfoDailyHandler)
	apiV2.GET("/get_diagnosis_days", GetDeviceDiagnosisDailyByDeviceIdHandler)
	// by-user_id or all node count
	apiV2.GET("/get_diagnosis_days_user", GetDeviceDiagnosisDailyByUserIdHandler)
	apiV2.GET("/get_diagnosis_hours", GetDeviceDiagnosisHourHandler)
	apiV2.GET("/get_cache_hours", GetCacheHourHandler)
	apiV2.GET("/get_cache_days", GetCacheDaysHandler)
	apiV2.GET("/get_applications", GetApplicationsHandler)
	apiV2.GET("/get_storage_stats", ListStorageStats)

	// node daily count
	apiV2.GET("/get_nodes_days", GetDiskDaysHandler)
	// console
	apiV2.GET("/device_update", DeviceUpdateHandler)
	// request from titan api
	apiV2.GET("/get_cache_list", GetCacheListHandler)
	apiV2.GET("/get_retrieval_list", GetRetrievalListHandler)
	apiV2.GET("/get_validation_list", GetValidationListHandler)
	apiV2.GET("/login_before", BeforeLogin)
	apiV2.POST("/login", authMiddleware.LoginHandler)
	apiV2.POST("/logout", authMiddleware.LogoutHandler)
	apiV2.GET("/get_user_device_count", GetUserDevicesCountHandler)

	apiV2.Use(authMiddleware.MiddlewareFunc())
	apiV2.Use(AuthRequired(authMiddleware))
	apiV2.GET("/get_device_info_auth", GetDeviceInfoHandler)
	apiV2.GET("/get_application_amount", GetApplicationAmountHandler)
	apiV2.POST("/create_application", CreateApplicationHandler)
	apiV2.GET("/device_binding", DeviceBindingHandler)
	apiV2.GET("/device_unbinding", DeviceUnBindingHandler)
	apiV2.GET("/get_user_device_profile", GetUserDeviceProfileHandler)
	apiV2.GET("/get_device_active_info", GetDeviceActiveInfoHandler)

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
	storage.GET("/get_map_info", GetMapInfoHandler)
	storage.POST("/get_verify_code", GetVerifyCodeHandle)
	storage.POST("/register", UserRegister)
	storage.POST("/password_reset", PasswordRest)
	storage.GET("/login_before", BeforeLogin)
	storage.POST("/login", authMiddleware.LoginHandler)
	storage.POST("/logout", authMiddleware.LogoutHandler)
	apiV3.GET("/", GetShareLinkHandler)
	storage.GET("/get_link", ShareLinkHandler)
	storage.GET("/get_map_cid", GetMapByCidHandler)
	storage.GET("/get_map_link", GetShareLinkHandler)
	storage.GET("/get_asset_detail", GetCarFileCountHandler)
	storage.GET("/get_asset_location", GetLocationHandler)
	storage.GET("/share_asset", ShareAssetsHandler)
	storage.GET("/get_asset_status", GetAssetStatusHandler)
	storage.GET("/get_fil_storage_list", GetFilStorageListHandler)
	storage.Use(authMiddleware.MiddlewareFunc())
	storage.Use(AuthRequired(authMiddleware))
	storage.GET("/get_locateStorage", GetAllocateStorageHandler)
	storage.GET("/get_storage_size", GetStorageSizeHandler)
	storage.GET("/get_vip_info", GetUserVipInfoHandler)
	storage.GET("/get_user_access_token", GetUserAccessTokenHandler)
	storage.GET("/create_asset", CreateAssetHandler)
	storage.GET("/delete_asset", DeleteAssetHandler)
	storage.GET("/get_asset_info", GetAssetInfoHandler)
	storage.GET("/get_asset_list", GetAssetListHandler)
	storage.GET("/get_all_asset_list", GetAssetAllListHandler)
	storage.GET("/share_status_set", UpdateShareStatusHandler)
	storage.GET("/create_key", CreateKeyHandler)
	storage.GET("/get_keys", GetKeyListHandler)
	storage.GET("/delete_key", DeleteKeyHandler)
	storage.GET("/get_asset_count", GetAssetCountHandler)
	storage.GET("/get_user_info_hour", GetStorageHourHandler)
	storage.GET("/get_user_info_daily", GetStorageDailyHandler)
	storage.GET("/refresh_token", authMiddleware.RefreshHandler)
	storage.GET("/new_secret", CreateNewSecretKeyHandler)
}

func RegisterV1RouterAuthByKey(router *gin.Engine) {
	authV1 := router.Group("/v1")
	storage := authV1.Group("/storage")
	storage.Use(AuthAPIKeyMiddlewareFunc())
	storage.POST("/add_fil_storage", CreateFilStorageHandler)
}
