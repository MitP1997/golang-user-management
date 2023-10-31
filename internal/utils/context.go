package utils

import (
	"github.com/MitP1997/golang-user-management/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetContextLogger(c *gin.Context, logger *zap.Logger) {
	c.Set("logger", logger)
}

func GetContextLogger(c *gin.Context) *zap.Logger {
	logger, ok := c.Get("logger")
	if !ok {
		return nil
	}
	return logger.(*zap.Logger)
}

func SetContextUser(c *gin.Context, user *models.User) {
	c.Set("user", user)
}

func GetContextUser(c *gin.Context) *models.User {
	user, ok := c.Get("user")
	if !ok {
		return nil
	}
	return user.(*models.User)
}

// Ideally we should create a new context and copy the values that are required.
// The reason for not taking the ideal approach is that the context.Context does not support Get and Set values directly,
// which is being used above in SetContextLogger and GetContextLogger.
func CreateDuplicateContext(c *gin.Context) *gin.Context {
	duplicateContext := c.Copy()
	return duplicateContext
}
