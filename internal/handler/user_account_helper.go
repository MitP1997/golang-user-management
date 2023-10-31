package handler

import (
	"fmt"

	"github.com/MitP1997/golang-user-management/internal/constants"
	"github.com/MitP1997/golang-user-management/internal/datatypes"
	"github.com/MitP1997/golang-user-management/internal/errors"
	"github.com/MitP1997/golang-user-management/internal/models"
	serviceRegistry "github.com/MitP1997/golang-user-management/internal/service_registry"
	"github.com/MitP1997/golang-user-management/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func sendEmailVerificationOtp(c *gin.Context, user *models.User) (err *errors.Error) {
	logger := utils.GetContextLogger(c)
	redisClient := serviceRegistry.GetRedisClient()
	otp, e := redisClient.SetUserTokenForScope(c, user.Id, constants.RedisUserEmailVerificationScope, constants.TokenType6DigitOtp)
	if e != nil {
		logger.Error("Error while generating user email verification token", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": err.Error()})
		return
	}
	mailBody := prepareMailBodyForEmailVerification(c, otp, user)

	// creating a duplicate context as the current context would die when the response ends
	// ideally we should create a new context with the required values instead of duplicating the current context
	dupCtx := utils.CreateDuplicateContext(c)
	// spawn a go routine to send the email as we do not want to add to the latency of the request
	go serviceRegistry.GetMailerClient().SendMail(dupCtx, user.Email, "Please verify your email", mailBody)
	return
}

func sendEmailChangePasswordOtp(c *gin.Context, user *models.User) (err *errors.Error) {
	logger := utils.GetContextLogger(c)
	redisClient := serviceRegistry.GetRedisClient()
	otp, e := redisClient.SetUserTokenForScope(c, user.Id, constants.RedisUserChangePasswordScope, constants.TokenType6DigitOtp)
	if e != nil {
		logger.Error("Error while generating user change password token", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": err.Error()})
		return
	}
	mailBody := prepareMailBodyForChangePassword(c, otp, user)

	// creating a duplicate context as the current context would die when the response ends
	// ideally we should create a new context with the required values instead of duplicating the current context
	dupCtx := utils.CreateDuplicateContext(c)
	// spawn a go routine to send the email as we do not want to add to the latency of the request
	go serviceRegistry.GetMailerClient().SendMail(dupCtx, user.Email, "OTP to change password", mailBody)
	return
}

func verifyUserPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func prepareMailBodyForEmailVerification(ctx *gin.Context, otp string, user *models.User) (mailBody string) {
	mailBody = fmt.Sprintf("Hello %s %s,\n\nThe OTP to verify your email address: %s\n\n%s", user.GivenName, user.FamilyName, otp, constants.MailSignature)
	return
}

func prepareMailBodyForChangePassword(ctx *gin.Context, otp string, user *models.User) (mailBody string) {
	mailBody = fmt.Sprintf("Hello %s %s,\n\nThis the OTP to change password: %s\n\n%s", user.GivenName, user.FamilyName, otp, constants.MailSignature)
	return
}

func verifyEmailOtp(ctx *gin.Context, otp string, scope datatypes.RedisScope, user *models.User) (err *errors.Error) {
	redisClient := serviceRegistry.GetRedisClient()
	token, e := redisClient.GetUserTokenForScope(ctx, user.Id, scope)
	if e != nil {
		if e.IsNotFound() {
			return errors.InvalidOtpError(nil)
		}
		return e
	}
	if otp != token {
		return errors.InvalidOtpError(nil)
	}
	return
}

func updateUserToMarkEmailVerified(ctx *gin.Context) (err *errors.Error) {
	user := utils.GetContextUser(ctx)
	if user == nil {
		return errors.UnauthedUserError(nil)
	}
	err = user.Update(ctx, bson.M{"_id": user.Id}, bson.M{"status": models.UserStatus_VERIFIED, "verified_at": timestamppb.Now()})
	return
}

func allowResendEmailOtp(ttl int32) bool {
	return ttl < (constants.RedisScopeTtl[constants.RedisUserEmailVerificationScope] - constants.RedisResendEmailAllowedAfter)
}
