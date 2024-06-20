package api

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/pkg/oss"
	"github.com/google/uuid"
)

func FileUploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(400, "Bad Request")
		return
	}
	if !isAllowedFileFormat(file) {
		c.String(400, "Unsupported file format")
		return
	}

	path := c.PostForm("path")
	if path == "" {
		c.String(400, "Bad Request")
		return
	}
	if !isAllowedPath(path) {
		c.String(400, "Unsupported path")
		return
	}

	newFileName := uuid.NewString() + filepath.Ext(file.Filename)
	// 根据日期创建文件夹
	path = fmt.Sprintf("%s/%s/%s", path, time.Now().Format("20060102"), newFileName)

	ossClient := oss.NewMustOssAPI(config.Cfg.Oss.EndPoint, config.Cfg.Oss.AccessId, config.Cfg.Oss.AccessKey)

	f, err := file.Open()
	if err != nil {
		c.String(400, "Invalid file")
		return
	}

	if err := ossClient.Upload(config.Cfg.Oss.Bucket, path, f); err != nil {
		log.Errorf("FileUploadHandler: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	f.Close()

	c.JSON(http.StatusOK, respJSON(map[string]string{"url": fmt.Sprintf("%s/%s", config.Cfg.Oss.Host, path)}))
}

func isAllowedFileFormat(file *multipart.FileHeader) bool {
	ext := filepath.Ext(file.Filename)
	return ext == ".jpeg" || ext == ".jpg" || ext == ".png" || ext == ".txt" || ext == ".json"
}

func isAllowedPath(path string) bool {
	return path == "ads" || path == "reports"
}
