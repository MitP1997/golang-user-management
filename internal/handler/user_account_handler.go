package handler

import (
	"net/http"
	"reflect"

	"github.com/MitP1997/golang-user-management/internal/constants"
	"github.com/MitP1997/golang-user-management/internal/errors"
	"github.com/MitP1997/golang-user-management/internal/models"
	"github.com/MitP1997/golang-user-management/internal/requests"
	"github.com/MitP1997/golang-user-management/internal/requirements"
	serviceRegistry "github.com/MitP1997/golang-user-management/internal/service_registry"
	"github.com/MitP1997/golang-user-management/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func GetUserSignupFormFields(c *gin.Context) {
	signupRequest := requests.SignupRequest{}
	value := reflect.ValueOf(signupRequest) // nolint:govet
	fields := map[string]interface{}{}
	numFields := value.NumField()
	structType := value.Type()

	for i := 0; i < numFields; i++ {
		field := structType.Field(i)

		formFieldType := field.Tag.Get("form_field_type")
		displayName := field.Tag.Get("display_name")
		if formFieldType != "" {
			formFieldDetails := map[string]string{
				"form_field_type": formFieldType,
				"display_name":    displayName,
			}
			fields[field.Tag.Get("form_field")] = formFieldDetails
		}
	}

	c.IndentedJSON(http.StatusOK, gin.H{"fields": fields})
}

func Signup(c *gin.Context) {
	logger := utils.GetContextLogger(c)

	var req requests.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Error while binding request body to user struct", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Password) < 8 {
		logger.Error("Password length less than 8 characters")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Password length less than 8 characters"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error while hashing password", zap.Error(err))
	}
	user := models.User{
		GivenName:  req.GivenName,
		FamilyName: req.FamilyName,
		Email:      req.Email,
		Password:   string(hashedPassword),
	}
	e := user.Insert(c)
	if e != nil {
		logger.Error("Error while inserting user into database", zap.Error(err))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	e = sendEmailVerificationOtp(c, &user)
	if e != nil {
		logger.Error("Error while sending email verification otp", zap.Error(err))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	logger = utils.AddKeyToContextLogger(c, "user_id", user.Id)
	redisClient := serviceRegistry.GetRedisClient()
	_, token, e := redisClient.GetOrCreateAndSetExpiryAuthToken(c, user.Id, "")
	if e != nil {
		logger.Error("Error while generating user auth token", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "User login successful", "token": token})
}

func Login(c *gin.Context) {
	logger := utils.GetContextLogger(c)

	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Error while binding request body to user struct", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"email": req.Email}
	var user models.User
	e := user.FindOne(c, filter)
	if e != nil {
		logger.Error("Error while fetching user from database", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}

	if verifyUserPassword(user.Password, req.Password) {
		logger = utils.AddKeyToContextLogger(c, "user_id", user.Id)
		redisClient := serviceRegistry.GetRedisClient()
		_, token, e := redisClient.GetOrCreateAndSetExpiryAuthToken(c, user.Id, "")
		if e != nil {
			logger.Error("Error while generating user auth token", zap.Error(e.Error()))
			c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "User login successful", "token": token})
		return
	}
	c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "User login password verification failed"})
}

func ResendEmailVerificationOtp(c *gin.Context) {
	logger := utils.GetContextLogger(c)

	user := utils.GetContextUser(c)
	if user.VerifiedAt != nil {
		logger.Error("User email already verified")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "User email already verified"})
		return
	}
	redisClient := serviceRegistry.GetRedisClient()
	ttl, e := redisClient.GetOtpTtl(c, user.Id)
	if e != nil {
		logger.Error("Error while fetching otp ttl", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	if !allowResendEmailOtp(int32(ttl.Seconds())) {
		logger.Error("Resend email otp not allowed")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Resend email otp not allowed"})
		return
	}
	e = sendEmailVerificationOtp(c, utils.GetContextUser(c))
	if e != nil {
		logger.Error("Error while sending email verification otp", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Email verification otp resent"})
}

func VerifyEmail(c *gin.Context) {
	logger := utils.GetContextLogger(c)

	var req requests.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Error while binding request body to user struct", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Otp == "" {
		logger.Error("OTP not present in request")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "OTP not present in request"})
		return
	}
	user := utils.GetContextUser(c)
	if user == nil {
		err := errors.UnauthedUserError(nil)
		logger.Error("Error while verifying email otp", zap.Error(err.Error()))
		c.IndentedJSON(err.UserStatusError(), gin.H{"error": err.UserErrorString()})
		return
	}
	e := verifyEmailOtp(c, req.Otp, constants.RedisUserEmailVerificationScope, user)
	if e != nil {
		logger.Error("Error while verifying email otp", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	if e := updateUserToMarkEmailVerified(c); e != nil {
		logger.Error("Error while updating user to mark email verified", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func GetPendingRequirements(c *gin.Context) {
	logger := utils.GetContextLogger(c)

	pendingRequirements := requirements.GetPendingRequirements(c)
	logger.Info("Pending requirements", zap.Any("pending_requirements", pendingRequirements))
	c.IndentedJSON(http.StatusOK, gin.H{"requirements": pendingRequirements})
}

func ChangePasswordInitiate(c *gin.Context) {
	logger := utils.GetContextLogger(c)

	var req requests.ChangePasswordInitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Error while binding request body to user struct", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email == "" {
		logger.Error("Email not present in request")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Email not provided"})
		return
	}

	filter := bson.M{"email": req.Email}
	var user models.User
	e := user.FindOne(c, filter)
	if e != nil {
		logger.Error("Error while fetching user from database", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	e = sendEmailChangePasswordOtp(c, &user)
	if e != nil {
		logger.Error("Error while sending change password otp", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
}

func ChangePassword(c *gin.Context) {
	logger := utils.GetContextLogger(c)

	var req requests.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Error while binding request body to user struct", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email == "" || req.Password == "" || req.Otp == "" {
		logger.Error("Email or Password or OTP not provided")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Email or Password or OTP not provided"})
		return
	}

	filter := bson.M{"email": req.Email}
	var user models.User
	e := user.FindOne(c, filter)
	if e != nil {
		logger.Error("Error while fetching user from database", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}

	e = verifyEmailOtp(c, req.Otp, constants.RedisUserChangePasswordScope, &user)
	if e != nil {
		logger.Error("Error while verifying email otp", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error while hashing password", zap.Error(err))
	}

	user.Password = string(hashedPassword)
	e = user.Update(c, bson.M{"_id": user.Id}, bson.M{"password": user.Password})
	if e != nil {
		logger.Error("Error while updating user password", zap.Error(e.Error()))
		c.IndentedJSON(e.UserStatusError(), gin.H{"error": e.UserErrorString()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
