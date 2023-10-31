package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getServerConfig() gin.HandlerFunc {
	config := cors.DefaultConfig()
	var allowOrigins []string
	allowOrigins = append(allowOrigins, strings.Split(os.Getenv("SERVER_ALLOW_ORIGINS"), ",")...)
	config.AllowOrigins = allowOrigins
	fmt.Println(allowOrigins)
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	return cors.New(config)
}
