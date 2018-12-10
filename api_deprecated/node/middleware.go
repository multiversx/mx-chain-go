package node

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(router *gin.Engine) {
	router.Use(cors.Default())
}
