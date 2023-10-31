package router

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	apiRouterGroup := r.Group("/api/v1")
	RegisterUserRoutes(apiRouterGroup)
}
