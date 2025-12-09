package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	exception "go-far/src/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

type Auth interface {
	GenerateToken(c *gin.Context, data any) (*TokenDetails, error)
	ValidateToken(c *gin.Context) (*AccessDetails, error)
	ValidateRefreshToken(c *gin.Context, token string) (*AccessDetails, error)
	// InvalidateToken(c *gin.Context) error
	// InvalidateRefreshToken(c *gin.Context, accessDetails *AccessDetails) error
	// CheckToken(c *gin.Context, userID string) error
}

var onceAuth = &sync.Once{}

type AuthOptions struct {
	PrivateKey          string        `yaml:"private_key"`
	PublicKey           string        `yaml:"public_key"`
	ExpiredToken        time.Duration `yaml:"expired_token"`
	ExpiredRefreshToken time.Duration `yaml:"expired_refresh_token"`
	// AutoLogoff          time.Duration `yaml:"auto_logoff"`
}

type auth struct {
	log                 zerolog.Logger
	redis               *redis.Client
	privateKey          []byte
	publicKey           []byte
	expiredToken        time.Duration
	expiredRefreshToken time.Duration
}

type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	AccessUUID   string `json:"-"`
	RefreshUUID  string `json:"-"`
	ExpiresAt    int64  `json:"expiresAt"`
	ExpiresRt    int64  `json:"expiresRt"`
}

type AccessDetails struct {
	AccessUUID  string
	RefreshUUID string
	UserID      string
	Username    string
}

func InitAuth(log zerolog.Logger, opt AuthOptions, redis *redis.Client) Auth {
	var a *auth

	onceAuth.Do(func() {
		privateKey, err := os.ReadFile(opt.PrivateKey)
		if err != nil {
			log.Panic().Err(err).Send()
		}

		publicKey, err := os.ReadFile(opt.PublicKey)
		if err != nil {
			log.Panic().Err(err).Send()
		}

		a = &auth{
			log:                 log,
			redis:               redis,
			privateKey:          privateKey,
			publicKey:           publicKey,
			expiredToken:        opt.ExpiredToken,
			expiredRefreshToken: opt.ExpiredRefreshToken,
		}
	})

	return a
}

func (a *auth) GenerateToken(c *gin.Context, data any) (*TokenDetails, error) {
	ctx := c.Request.Context()

	td := &TokenDetails{}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(a.privateKey)
	if err != nil {
		return nil, exception.InternalServerError("Failed to parse key")
	}

	dataVal := reflect.ValueOf(data)
	publicID := dataVal.FieldByName("PublicID").String()
	username := dataVal.FieldByName("Username").String()

	td.ExpiresAt = time.Now().Add(a.expiredToken).Unix()
	td.AccessUUID = ksuid.New().String()

	td.ExpiresRt = time.Now().Add(a.expiredRefreshToken).Unix()
	td.RefreshUUID = td.AccessUUID + "++" + publicID

	at := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"exp":         td.ExpiresAt,
		"access_uuid": td.AccessUUID,
		"user_id":     publicID,
		"name":        username,
		"authorized":  true,
	})

	td.AccessToken, err = at.SignedString(key)
	if err != nil {
		return nil, exception.InternalServerError("Failed to sign access token")
	}

	rt := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"exp":          td.ExpiresRt,
		"refresh_uuid": td.RefreshUUID,
		"user_id":      publicID,
		"name":         username,
	})

	td.RefreshToken, err = rt.SignedString(key)
	if err != nil {
		return nil, exception.InternalServerError("Failed to sign refresh token")
	}

	err = a.saveToRedis(ctx, publicID, td)
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (a *auth) saveToRedis(ctx context.Context, publicID string, td *TokenDetails) error {
	respAccess := a.redis.Set(ctx, td.AccessUUID, publicID, a.expiredToken)
	if respAccess.Err() != nil {
		return exception.InternalServerError("Failed to store access token in Redis")
	}

	respRefresh := a.redis.Set(ctx, td.RefreshUUID, publicID, a.expiredRefreshToken)
	if respRefresh.Err() != nil {
		return exception.InternalServerError("Failed to store refresh token in Redis")
	}

	return nil
}

func (a *auth) ValidateToken(c *gin.Context) (*AccessDetails, error) {
	return a.checkingToken(c)
}

func (a *auth) checkingToken(c *gin.Context) (*AccessDetails, error) {
	ctx := c.Request.Context()

	tokenStr := a.extractToken(c)
	token, err := a.verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, exception.UnauthorizedError("Invalid token")
	}

	userID := claims["user_id"].(string)
	username := claims["name"].(string)

	var accessUUID, redisIDUser string

	accessUUID, ok = claims["access_uuid"].(string)
	if !ok {
		return nil, exception.UnauthorizedError("Failed claims accessUUID")
	}

	redisIDUser, err = a.redis.Get(ctx, accessUUID).Result()
	if err != nil {
		return nil, exception.UnauthorizedError("Failed to get token from Redis")
	}

	if userID != redisIDUser {
		return nil, exception.UnauthorizedError("Authentication failure")
	}

	return &AccessDetails{
		AccessUUID: accessUUID,
		UserID:     redisIDUser,
		Username:   username,
	}, nil
}

func (a *auth) extractToken(c *gin.Context) string {
	bearToken := c.Request.Header["Authorization"][0]
	if len(bearToken) == 0 {
		return ""
	}

	tokenArr := strings.Split(bearToken, " ")
	if len(tokenArr) == 2 {
		return tokenArr[1]
	}

	return ""
}

func (a *auth) verifyToken(tokenStr string) (*jwt.Token, error) {
	key, err := jwt.ParseRSAPublicKeyFromPEM(a.publicKey)
	if err != nil {
		return nil, exception.InternalServerError("Failed to parse key")
	}

	token, err := jwt.Parse(tokenStr, func(jwtToken *jwt.Token) (any, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, exception.InternalServerError(fmt.Sprintf("unexpected signing method: %v", jwtToken.Header["alg"]))
		}

		return key, nil
	})
	if err != nil {
		return nil, exception.InternalServerError("Failed to parse token")
	}

	return token, nil
}

func (a *auth) ValidateRefreshToken(c *gin.Context, tokenStr string) (*AccessDetails, error) {
	ctx := c.Request.Context()

	token, err := a.verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, exception.UnauthorizedError("Invalid token")
	}

	userID := claims["user_id"].(string)
	username := claims["name"].(string)

	var accessUUID, refreshUUID, redisIDUser string

	refreshUUID, ok = claims["refresh_uuid"].(string)
	if !ok {
		return nil, err
	}

	redisIDUser, err = a.redis.Get(ctx, refreshUUID).Result()
	if err != nil {
		return nil, exception.InternalServerError("Failed to get token from Redis")
	}

	if userID != redisIDUser {
		return nil, exception.UnauthorizedError("Authentication failure")
	}

	return &AccessDetails{
		AccessUUID:  accessUUID,
		RefreshUUID: refreshUUID,
		UserID:      redisIDUser,
		Username:    username,
	}, nil
}
