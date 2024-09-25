package storage

import "time"

// UserAccessControl 用户文件控制
type userAccessControl = string

const (
	userAPIKeyReadFile     userAccessControl = "readFile"
	userAPIKeyCreateFile   userAccessControl = "createFile"
	userAPIKeyDeleteFile   userAccessControl = "deleteFile"
	userAPIKeyReadFolder   userAccessControl = "readFolder"
	userAPIKeyCreateFolder userAccessControl = "createFolder"
	userAPIKeyDeleteFolder userAccessControl = "deleteFolder"
)

var userAccessControlAll = []userAccessControl{
	userAPIKeyReadFile,
	userAPIKeyCreateFile,
	userAPIKeyDeleteFile,
	userAPIKeyReadFolder,
	userAPIKeyCreateFolder,
	userAPIKeyDeleteFolder,
}

var funcAccessControlMap = map[string]userAccessControl{
	"CreateAsset":      userAPIKeyCreateFile,
	"ListAssets":       userAPIKeyReadFile,
	"DeleteAsset":      userAPIKeyDeleteFile,
	"ShareAssets":      userAPIKeyReadFile,
	"CreateAssetGroup": userAPIKeyCreateFolder,
	"ListAssetGroup":   userAPIKeyReadFolder,
	"DeleteAssetGroup": userAPIKeyDeleteFolder,
	"RenameAssetGroup": userAPIKeyCreateFolder,
}

// UserAPIKeysInfo 用户 api key 信息
type UserAPIKeysInfo struct {
	CreatedTime time.Time
	APIKey      string
}

// UserAPIKeySecretInfo 用户 api key secret信息
type UserAPIKeySecretInfo struct {
	APIKey      string
	APISecret   string
	CreatedTime time.Time
}

// UserKeyInfo 用户key信息
type UserKeyInfo struct {
	UID  string
	Salt string
}

// TenantInfo 租户信息
type TenantInfo struct {
	TenantID string
	Name     string
	Salt     string
}
