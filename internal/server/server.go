package server

import (
	"github.com/MitP1997/golang-user-management/internal/middleware"
	"github.com/MitP1997/golang-user-management/internal/models"
	"github.com/MitP1997/golang-user-management/internal/router"
	serviceRegistry "github.com/MitP1997/golang-user-management/internal/service_registry"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func NewServer() *gin.Engine {
	err := loadEnv()
	if err != nil {
		panic(err)
	}

	serviceRegistry.InitServiceRegistry()
	err = models.InitMongoConnection()
	if err != nil {
		panic(err)
	}
	r := gin.Default()

	r.Use(getServerConfig())

	r.Use(middleware.IntroduceLoggingContextMiddleware(serviceRegistry.GetLogger()))
	r.Use(middleware.GinzapLoggerMiddleware())
	r.Use(middleware.RecoveryWithZapMiddleware())

	router.RegisterRoutes(r)
	return r
}

func loadEnv() (err error) {
	err = godotenv.Load()
	return
}
