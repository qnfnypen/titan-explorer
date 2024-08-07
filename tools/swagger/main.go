package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	r := gin.New()
	r.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("../../docs/swagger.json")))
	r.Run(":8080")
}
