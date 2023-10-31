package middleware

import (
	"time"

	"github.com/MitP1997/golang-user-management/internal/utils"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func IntroduceLoggingContextMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// generate unique request id and add to context
		// request id as a uuid
		requestId := uuid.New().String()
		requestIdCoreField := zapcore.Field{Key: "request_id", Type: zapcore.StringType, String: requestId}
		logger := logger.With(requestIdCoreField)
		utils.SetContextLogger(c, logger)
		c.Next()
	}
}

func GinzapLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetContextLogger(c)
		ginzap.Ginzap(logger, time.RFC3339, true)(c)
	}
}

func RecoveryWithZapMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetContextLogger(c)
		ginzap.RecoveryWithZap(logger, true)(c)
	}
}
