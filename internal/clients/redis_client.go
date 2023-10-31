package clients

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/MitP1997/golang-user-management/internal/constants"
	"github.com/MitP1997/golang-user-management/internal/datatypes"
	"github.com/MitP1997/golang-user-management/internal/errors"
	"github.com/MitP1997/golang-user-management/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	client *redis.Client
}

type RedisKey struct {
	Key   string
	Scope datatypes.RedisScope
}

func (r *RedisKey) String() string {
	return fmt.Sprintf("%s_|_%s", r.Scope, r.Key)
}

func NewRedisClient() *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		}),
	}
}

func (r *RedisClient) SetWithExpiration(ctx context.Context, key RedisKey, value interface{}, expiration int32) (err *errors.Error) {
	e := r.client.Set(ctx, key.String(), value, time.Duration(expiration)*time.Second).Err()
	if e != nil {
		return errors.RedisInternalServerError(e)
	}
	return
}

func (r *RedisClient) Get(ctx context.Context, key RedisKey) (val interface{}, err *errors.Error) {
	val, e := r.client.Get(ctx, key.String()).Result()
	if e != nil {
		if e == redis.Nil {
			return nil, errors.RedisNotFoundError(e)
		}
		return nil, errors.RedisInternalServerError(e)
	}
	return
}

func (r *RedisClient) GetTtl(ctx context.Context, key RedisKey) (ttl time.Duration, err *errors.Error) {
	ttl, e := r.client.TTL(ctx, key.String()).Result()
	if e != nil {
		if e == redis.Nil {
			return 0, errors.RedisNotFoundError(e)
		}
		return 0, errors.RedisInternalServerError(e)
	}
	return
}

func (r *RedisClient) GetAndResetExpiration(ctx context.Context, key RedisKey, expiration int32) (val interface{}, err *errors.Error) {
	val, e := r.client.Get(ctx, key.String()).Result()
	if e != nil {
		if e == redis.Nil {
			return nil, errors.RedisNotFoundError(e)
		}
		return nil, errors.RedisInternalServerError(e)
	}
	e = r.client.Expire(ctx, key.String(), time.Duration(expiration)*time.Second).Err()
	if e != nil {
		return nil, errors.RedisInternalServerError(e)
	}
	return
}

func (r *RedisClient) SetUserTokenForScope(ctx *gin.Context, userId string, scope datatypes.RedisScope, tokenType datatypes.TokenType) (token string, err *errors.Error) {
	logger := utils.GetContextLogger(ctx)
	logger = logger.With(zap.String("scope", string(scope)))
	token = generateToken(tokenType)
	err = r.SetWithExpiration(ctx, RedisKey{Key: userId, Scope: scope}, token, getRedisScopeTtl(scope))
	if err != nil {
		logger.Error("Error while setting token in redis", zap.Error(err.Error()))
		return "", err
	}
	return
}

func (r *RedisClient) GetUserTokenForScope(ctx *gin.Context, userId string, scope datatypes.RedisScope) (token string, err *errors.Error) {
	logger := utils.GetContextLogger(ctx)
	logger = logger.With(zap.String("scope", string(scope)))
	val, err := r.Get(ctx, RedisKey{Key: userId, Scope: scope})
	if err != nil {
		logger.Error("Error while getting token from redis", zap.Error(err.Error()))
		return
	}
	token = val.(string)
	return
}

// common function that generates and refreshes both the keys i.e. user key and token key in redis at the same time
func (r *RedisClient) GetOrCreateAndSetExpiryAuthToken(ctx *gin.Context, userId string, token string) (string, string, *errors.Error) {
	logger := utils.GetContextLogger(ctx)
	var userRedisKey, tokenRedisKey RedisKey

	if token == "" && userId == "" {
		logger.Info("Both token and userId are empty")
		return "", "", errors.MissingUserIdAndTokenError()
	}
	if token == "" {
		userRedisKey = RedisKey{Key: userId, Scope: constants.RedisUserAuthTokenScope}
		tokenVal, _ := r.Get(ctx, userRedisKey)
		if tokenVal != nil {
			token = tokenVal.(string)
		}
	}
	// Ideally we don't need this check below
	if userId == "" {
		tokenRedisKey = RedisKey{Key: token, Scope: constants.RedisAuthTokenUserScope}
		userIdVal, _ := r.Get(ctx, tokenRedisKey)
		if userIdVal != nil {
			userId = userIdVal.(string)
		}
	}

	pipe := r.client.Pipeline()
	if token == "" {
		token = generateToken(constants.TokenTypeUuid)
	}

	tokenRedisKey = RedisKey{Key: token, Scope: constants.RedisAuthTokenUserScope}
	userRedisKey = RedisKey{Key: userId, Scope: constants.RedisUserAuthTokenScope}
	pipe.Set(ctx, userRedisKey.String(), token, time.Duration(getRedisScopeTtl(constants.RedisUserAuthTokenScope))*time.Second)
	pipe.Set(ctx, tokenRedisKey.String(), userId, time.Duration(getRedisScopeTtl(constants.RedisAuthTokenUserScope))*time.Second)

	_, e := pipe.Exec(ctx)
	if e != nil {
		logger.Error("Error while getting auth token from redis", zap.Error(e))
		return "", "", errors.RedisInternalServerError(e)
	}
	return userId, token, nil
}

func (r *RedisClient) GetOtpTtl(ctx *gin.Context, userId string) (ttl time.Duration, err *errors.Error) {
	ttl, e := r.GetTtl(ctx, RedisKey{Key: userId, Scope: constants.RedisUserEmailVerificationScope})
	if e != nil {
		return 0, e
	}
	return
}

func getRedisScopeTtl(scope datatypes.RedisScope) int32 {
	return constants.RedisScopeTtl[scope]
}
