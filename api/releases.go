package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/google/go-github/v60/github"
	"net/http"
	"strings"
	"time"
)

type Release struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
}

func GetReleasesHandler(c *gin.Context) {
	client := github.NewClient(nil)

	release, err := GetReleaseFromCache(c.Request.Context())
	if err == nil {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"release": release,
		}))
		return
	}

	releases, _, err := client.Repositories.GetLatestRelease(context.Background(), "Titannet-dao", "titan-node")
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoBearerToken, c))
		return
	}

	var out []*Release
	for _, release := range releases.Assets {
		if strings.HasSuffix(*release.Name, "sha256") {
			continue
		}

		var platform string
		if strings.Contains(*release.Name, "darwin") {
			platform = "macOS"
		}
		if strings.Contains(*release.Name, "windows") {
			platform = "Windows"
		}
		if strings.Contains(*release.Name, "linux") {
			platform = "Linux"
		}

		var arch string
		splits := strings.Split(*release.Name, "_")
		if len(splits) > 3 {
			arch = strings.ToUpper(strings.Split(splits[3], ".")[0])
		}

		out = append(out, &Release{
			OS:          platform,
			Arch:        arch,
			Version:     *releases.TagName,
			DownloadURL: *release.BrowserDownloadURL,
		})
	}

	err = CacheRelease(c.Request.Context(), out)
	if err != nil {
		log.Errorf("cache release: %v", err)
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"release": out,
	}))
}

func CacheRelease(ctx context.Context, info []*Release) error {
	key := fmt.Sprintf("TITAN::RELEASE")

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	expiration := time.Minute * 5
	_, err = dao.RedisCache.Set(ctx, key, data, expiration).Result()
	if err != nil {
		log.Errorf("set release info: %v", err)
	}

	return nil
}

func GetReleaseFromCache(ctx context.Context) ([]*Release, error) {
	key := fmt.Sprintf("TITAN::RELEASE")
	result, err := dao.RedisCache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var out []*Release
	err = json.Unmarshal([]byte(result), &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
