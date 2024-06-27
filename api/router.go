package api

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("api")

func RegisterRouters(route *gin.Engine, cfg config.Config) {
	RegisterRouterWithJWT(route, cfg)
	RegisterRouterWithAPIKey(route)
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		log.Debug(string(body))
		//log.Debug(c.Request.Header)
		c.Next()
	}
}

func RegisterRouterWithJWT(router *gin.Engine, cfg config.Config) {
	apiV1 := router.Group("/api/v1")
	apiV2 := router.Group("/api/v2")
	link := router.Group("/link")

	authMiddleware, err := jwtGinMiddleware(cfg.SecretKey)
	if err != nil {
		log.Fatalf("jwt auth middleware: %v", err)
	}

	err = authMiddleware.MiddlewareInit()
	if err != nil {
		log.Fatalf("authMiddleware.MiddlewareInit: %v", err)
	}

	// testnet
	apiV2.POST("/subscribe", SubscribeHandler)

	// dashboard
	// Deprecated: use /height instead
	apiV2.GET("/get_high", GetBlockHeightHandler)
	apiV2.GET("/height", GetBlockHeightHandler)
	apiV2.GET("/all_areas", GetAllAreas)
	apiV2.GET("/get_index_info", GetIndexInfoHandler)
	apiV2.GET("/get_query_info", GetQueryInfoHandler)
	apiV2.GET("/releases", GetReleasesHandler)
	apiV2.GET("/app_version", GetAppVersionHandler)
	apiV2.POST("/device", GetDeviceProfileHandler)
	apiV2.POST("/device/binding", DeviceBindingHandler)
	apiV2.GET("/device/query_code", QueryDeviceCodeHandler)
	apiV2.GET("/device/distribution", GetDeviceDistributionHandler)
	apiV2.POST("/data/collection", DataCollectionHandler)

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
	//apiV2.GET("/get_applications", GetApplicationsHandler)
	apiV2.GET("/get_storage_stats", ListStorageStats)

	// node daily count
	apiV2.GET("/get_nodes_days", GetDiskDaysHandler)
	// console
	apiV2.GET("/device_update", DeviceUpdateHandler)
	// request from titan api
	apiV2.GET("/get_cache_list", GetCacheListHandler)
	apiV2.GET("/get_retrieval_list", GetRetrievalListHandler)
	apiV2.GET("/get_validation_list", GetValidationListHandler)
	apiV2.GET("/get_replica_list", GetReplicaListHandler)
	apiV2.GET("/get_profit_details", GetProfitDetailsHandler)
	apiV2.GET("/login_before", GetNonceStringHandler)
	apiV2.POST("/login", authMiddleware.LoginHandler)
	apiV2.POST("/logout", authMiddleware.LogoutHandler)
	apiV2.GET("/get_user_device_count", GetUserDevicesCountHandler)

	apiV2.Use(authMiddleware.MiddlewareFunc())
	apiV2.Use(AuthRequired(authMiddleware))
	apiV2.GET("/get_device_info_auth", GetDeviceInfoHandler)
	apiV2.GET("/device_unbinding", DeviceUnBindingHandlerOld)
	apiV2.GET("/get_user_device_profile", GetUserDeviceProfileHandler)
	apiV2.GET("/get_device_active_info", GetDeviceActiveInfoHandler)
	apiV2.POST("/wallet/bind", BindWalletHandler)
	apiV2.POST("/wallet/unbind", UnBindWalletHandler)
	apiV2.GET("/referral_list", GetReferralListHandler)
	apiV2.GET("/generate/code", GenerateCodeHandler)

	// user
	user := apiV1.Group("/user")
	user.POST("/register", UserRegister)
	user.POST("/password_reset", PasswordRest)
	user.GET("/verify_code", GetNumericVerifyCodeHandler)
	user.POST("/login", authMiddleware.LoginHandler)
	user.POST("/logout", authMiddleware.LogoutHandler)
	user.GET("/ads/banners", GetBannersHandler)
	user.GET("/ads/notices", GetNoticesHandler)
	user.POST("/upload", FileUploadHandler)
	user.POST("/bugs/report", BugReportHandler)
	user.GET("/bugs/list", MyBugReportListHandler)
	user.Use(authMiddleware.MiddlewareFunc())
	user.GET("/refresh_token", authMiddleware.RefreshHandler)
	user.POST("/info", GetUserInfoHandler)
	user.POST("/referral_code/new", AddReferralCodeHandler)
	user.GET("/referral_code/detail", GetReferralCodeDetailHandler)
	user.GET("/referral_code/stat", GetReferralCodeStatHandler)

	// admin
	admin := apiV1.Group("/admin")
	adminMiddleware := *authMiddleware
	adminMiddleware.Authorizator = AdminOnly
	err = adminMiddleware.MiddlewareInit()
	if err != nil {
		log.Fatal(err)
	}

	admin.Use(adminMiddleware.MiddlewareFunc())
	admin.GET("/get_login_log", GetLoginLogHandler)
	admin.GET("/get_operation_log", GetOperationLogHandler)
	admin.GET("/get_node_daily_trend", GetNodeDailyTrendHandler)
	admin.GET("/kol/list", GetKOLListHandler)
	admin.POST("/kol/add", AddKOLHandler)
	admin.POST("/kol/update", UpdateKOLHandler)
	admin.POST("/kol/delete", DeleteKOLHandler)
	admin.POST("/kol_level/add", AddKOLLevelHandler)
	admin.GET("/kol_level/list", GetKOLLevelConfigHandler)
	admin.POST("/kol_level/update", UpdateKOLLevelHandler)
	admin.POST("/kol_level/delete", DeleteKOLLevelHandler)
	admin.GET("/referral_reward_daily", GetReferralRewardDailyHandler)
	admin.GET("/referral_reward_daily/export", ExportReferralRewardDailyHandler)
	// ads
	admin.GET("/ads/list", ListAdsHandler)
	admin.POST("/ads/add", AddAdsHandler)
	admin.POST("/ads/delete", DeleteAdsHandler)
	admin.POST("/ads/update", UpdateAdsHandler)
	admin.POST("/upload", FileUploadHandler)
	// bugs
	admin.GET("/bugs/list", BugReportListHandler)
	admin.POST("/bugs/edit", BugEditHandler)

	// storage
	storage := apiV1.Group("/storage")
	storage.GET("/get_map_info", GetMapInfoHandler)
	// Deprecated: use /user/verify_code instead
	storage.POST("/get_verify_code", GetNumericVerifyCodeHandler)
	// Deprecated: use /user/register instead
	storage.POST("/register", UserRegister)
	// Deprecated: use /user/password_reset instead
	storage.POST("/password_reset", PasswordRest)
	storage.GET("/login_before", GetNonceStringHandler)
	storage.POST("/login", authMiddleware.LoginHandler)
	storage.POST("/logout", authMiddleware.LogoutHandler)
	link.GET("/", GetShareLinkHandler)
	storage.GET("/get_link", ShareLinkHandler)
	storage.GET("/get_map_cid", GetMapByCidHandler)
	storage.GET("/get_map_link", GetShareLinkHandler)
	storage.GET("/get_asset_detail", GetAssetDetailHandler)
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
	storage.GET("/get_upload_info", GetUploadInfoHandler)
	storage.GET("/create_asset", CreateAssetHandler)
	storage.POST("/create_asset", CreateAssetPostHandler)
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
	storage.GET("/get_key_perms", GetAPIKeyPermsHandler)
	storage.GET("/create_group", CreateGroupHandler)
	storage.GET("/get_groups", GetGroupsHandler)
	storage.GET("/get_asset_group_list", GetAssetGroupListHandler)
	storage.GET("/delete_group", DeleteGroupHandler)
	storage.GET("/rename_group", RenameGroupHandler)
	storage.GET("/move_group_to_group", MoveGroupToGroupHandler)
	storage.GET("/move_asset_to_group", MoveAssetToGroupHandler)

	container := apiV1.Group("/container")
	container.Use(authMiddleware.MiddlewareFunc())
	container.GET("/providers", GetProvidersHandler)
	container.GET("/deployments", GetDeploymentsHandler)
	container.GET("/deployment/manifest", GetDeploymentManifestHandler)
	container.POST("/deployment/create", CreateDeploymentHandler)
	container.POST("/deployment/delete", DeleteDeploymentHandler)
	container.POST("/deployment/update", UpdateDeploymentHandler)
	container.GET("/deployment/logs", GetDeploymentLogsHandler)
	container.GET("/deployment/event", GetDeploymentEventsHandler)
	container.GET("/deployment/domains", GetDeploymentDomainHandler)
	container.POST("/deployment/domain/add", AddDeploymentDomainHandler)
	container.POST("/deployment/domain/del", DeleteDeploymentDomainHandler)
	container.GET("/deployment/shell", GetDeploymentShellHandler)

	//signature
	signature := apiV1.Group("/sign")
	signature.GET("/info", getSignInfo)
	signature.GET("/summary", getSummaryInfo)

	signature.POST("/upload", setSignInfo)
	signature.POST("/command", getCommand)

	// url
	uri := apiV1.Group("/url")
	uri.GET("/discord", getDiscordURL)

	// test1
	test1 := apiV1.Group("/test1")
	test1.Use(authMiddleware.MiddlewareFunc())
	tnc := &Test1NodeController{}
	// test1.GET("/node/info", tnc.GetNodes)
	// test1.PUT("/node/update_name", tnc.UpdateDeviceName)
	test1.GET("/node/nums", tnc.GetNodeNums)
	test1.PUT("/node/delete_offline", tnc.DeleteOffLineNode)
	test1.PUT("/node/move_back_deleted", tnc.MoveBackDeletedNode)
}

func RegisterRouterWithAPIKey(router *gin.Engine) {
	authV1 := router.Group("/v1")
	storage := authV1.Group("/storage")
	storage.Use(AuthAPIKeyMiddlewareFunc())
	storage.POST("/add_fil_storage", CreateFilStorageHandler)
	storage.GET("/backup_assets", GetBackupAssetsHandler)
	storage.POST("/backup_result", BackupResultHandler)

	app := authV1.Group("/app")
	app.Use(AuthAPIKeyMiddlewareFunc())
	app.POST("/new_version", CreateAppVersionHandler)
}
