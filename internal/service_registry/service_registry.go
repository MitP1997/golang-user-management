package server

import (
	"github.com/MitP1997/golang-user-management/internal/clients"
	"go.uber.org/zap"
)

var serviceRegistry map[string]interface{}

const (
	keyLogger       = "key-logger"
	keyRedisClient  = "key-redis-client"
	keyMailerClient = "key-mailer-client"
)

func InitServiceRegistry() {
	serviceRegistry = make(map[string]interface{})
	logger, _ := zap.NewProduction()
	redisClient := clients.NewRedisClient()
	mailerClient := clients.NewMailerClient(logger)
	addClientToServiceRegistry(keyLogger, logger)
	addClientToServiceRegistry(keyRedisClient, redisClient)
	addClientToServiceRegistry(keyMailerClient, mailerClient)
}

func GetLogger() *zap.Logger {
	return getClientFromServiceRegistry(keyLogger).(*zap.Logger)
}

func GetRedisClient() *clients.RedisClient {
	return getClientFromServiceRegistry(keyRedisClient).(*clients.RedisClient)
}

func GetMailerClient() *clients.MailerClient {
	return getClientFromServiceRegistry(keyMailerClient).(*clients.MailerClient)
}

func addClientToServiceRegistry(key string, value interface{}) {
	serviceRegistry[key] = value
}

func getClientFromServiceRegistry(key string) interface{} {
	return serviceRegistry[key]
}
