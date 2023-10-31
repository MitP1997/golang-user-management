package handler

import (
	"net/http"

	"github.com/MitP1997/golang-user-management/internal/models"
	serviceRegistry "github.com/MitP1997/golang-user-management/internal/service_registry"
	"github.com/MitP1997/golang-user-management/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// check if authorization headers are present
//
//	if present, add the user object to context and continue
//	else return unauthorized error in response
func IsAuthorized(fn gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetContextLogger(c)
		token := c.GetHeader("Authorization")
		if token == "" {
			logger.Info("No Authorization headers provided")
			c.IndentedJSON(http.StatusForbidden, gin.H{"message": "No Authorization headers provided"})
			return
		}
		redisClient := serviceRegistry.GetRedisClient()
		userId, _, err := redisClient.GetOrCreateAndSetExpiryAuthToken(c, "", token)
		if err != nil || userId == "" {
			logger.Error("Error in GetOrCreateAndSetExpiryAuthToken or userId is empty", zap.Error(err.Error()))
			c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Please login!"})
			return
		}

		var user models.User
		e := user.FindOne(c, bson.M{"_id": userId})
		if e != nil {
			logger.Error("Error while finding user in database", zap.Error(e.Error()))
			c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
			return
		}
		utils.SetContextUser(c, &user)
		fn(c)
	}
}
