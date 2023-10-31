package router

import (
	"github.com/MitP1997/golang-user-management/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup) {
	userRouterGroup := r.Group("/user")
	userRouterGroup.GET("/signup-form-fields", handler.GetUserSignupFormFields)
	userRouterGroup.POST("/signup", handler.Signup)
	userRouterGroup.POST("/login", handler.Login)
	userRouterGroup.POST("/verify-email", handler.IsAuthorized(handler.VerifyEmail))
	userRouterGroup.POST("/resend-email-verification-otp", handler.IsAuthorized(handler.ResendEmailVerificationOtp))
	userRouterGroup.GET("/pending-requirements", handler.IsAuthorized(handler.GetPendingRequirements))
	userRouterGroup.POST("/change-password-initiate", handler.ChangePasswordInitiate)
	userRouterGroup.POST("/change-password", handler.ChangePassword)
}
