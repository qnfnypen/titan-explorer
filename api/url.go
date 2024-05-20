package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
)

func getDiscordURL(c *gin.Context) {
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"url": config.Cfg.URL.Discord,
	}))
}
