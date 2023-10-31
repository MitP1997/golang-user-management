package constants

import "github.com/MitP1997/golang-user-management/internal/datatypes"

const (
	RedisUserAuthTokenScope         datatypes.RedisScope = "user_auth_token"
	RedisAuthTokenUserScope         datatypes.RedisScope = "auth_token_user"
	RedisUserEmailVerificationScope datatypes.RedisScope = "user_email_verification"
	RedisUserChangePasswordScope    datatypes.RedisScope = "user_change_password"

	// allowing resend email verification otp after 5 mins
	RedisResendEmailAllowedAfter = 30
)

var RedisScopeTtl = map[datatypes.RedisScope]int32{
	// ttl of 1 day
	// TODO: Tweak the below params or may be fetch fro config/env
	RedisUserAuthTokenScope:      86400,
	RedisAuthTokenUserScope:      86400,
	RedisUserChangePasswordScope: 86400,
	// ttl of 15 mins
	RedisUserEmailVerificationScope: 900,
}
