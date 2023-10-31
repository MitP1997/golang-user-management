package requirements

import (
	"github.com/MitP1997/golang-user-management/internal/models"
	"github.com/MitP1997/golang-user-management/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetPendingRequirements(c *gin.Context) []string {
	user := utils.GetContextUser(c)
	requirements := []string{}
	if user.Status == models.UserStatus_UNVERIFIED {
		requirements = append(requirements, "email")
	}
	return requirements
}
