package token

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	appErr "go-far/internal/model/errors"
	"go-far/internal/preference"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

const (
	refreshTokenRotation   = true
	minSecretLength        = 32
	redisTimeout           = 3 * time.Second
	RefreshTokenUsedPrefix = "rt_used:"
)

type token struct {
	log                 *zerolog.Logger
	redis               *redis.Client
	secret              []byte
	expiredToken        time.Duration
	expiredRefreshToken time.Duration
}

var (
	onceToken = &sync.Once{}
	tokenInst *token
)

// Token defines the token management interface
type Token interface {
	GenerateToken(r *http.Request, data any) (*TokenDetails, error)
	ValidateToken(r *http.Request) (*AccessDetails, error)
	ValidateRefreshToken(r *http.Request, token string) (*AccessDetails, error)
}

// TokenOptions holds token configuration
type TokenOptions struct {
	ExpiredToken        time.Duration `yaml:"expired_token"`
	ExpiredRefreshToken time.Duration `yaml:"expired_refresh_token"`
}

// TokenDetails holds token information
type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	AccessUUID   string `json:"-"`
	RefreshUUID  string `json:"-"`
	ExpiresAt    int64  `json:"expiresAt"`
	ExpiresRt    int64  `json:"expiresRt"`
}

// AccessDetails holds access token details
type AccessDetails struct {
	AccessUUID  string
	RefreshUUID string
	UserID      string
	Username    string
	Role        string
}

// InitToken initializes the token module
func InitToken(log *zerolog.Logger, opt *TokenOptions, redisClient *redis.Client) Token {
	onceToken.Do(func() {
		secret := os.Getenv("JWT_SECRET_GO_FAR")
		if secret == "" {
			log.Panic().Msgf("Environment variable %s is not set", "JWT_SECRET_GO_FAR")
		}

		if len(secret) < minSecretLength {
			log.Panic().Msgf("JWT secret must be at least %d characters", minSecretLength)
		}

		tokenInst = &token{
			log:                 log,
			redis:               redisClient,
			secret:              []byte(secret),
			expiredToken:        opt.ExpiredToken,
			expiredRefreshToken: opt.ExpiredRefreshToken,
		}
	})

	return tokenInst
}

func (a *token) GenerateToken(r *http.Request, data any) (*TokenDetails, error) {
	ctx := r.Context()
	td := &TokenDetails{}
	var err error

	dataVal := reflect.ValueOf(data)
	if dataVal.Kind() == reflect.Ptr {
		dataVal = dataVal.Elem()
	}

	if !dataVal.IsValid() {
		return nil, appErr.NewWithCode(appErr.CodeHTTPBadRequest, "Invalid data for token generation")
	}

	publicIDField := dataVal.FieldByName("PublicID")
	usernameField := dataVal.FieldByName("Username")
	roleField := dataVal.FieldByName("Role")

	if !publicIDField.IsValid() || !usernameField.IsValid() || !roleField.IsValid() {
		return nil, appErr.NewWithCode(appErr.CodeHTTPBadRequest, "Data must contain PublicID, Username and Role fields")
	}

	publicID := publicIDField.String()
	username := usernameField.String()
	role := roleField.String()

	if publicID == "" || username == "" || role == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPBadRequest, "PublicID, Username and Role cannot be empty")
	}

	td.ExpiresAt = time.Now().Add(a.expiredToken).Unix()
	td.AccessUUID = ksuid.New().String()

	td.ExpiresRt = time.Now().Add(a.expiredRefreshToken).Unix()
	td.RefreshUUID = td.AccessUUID + preference.TokenSeparator + publicID

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":         td.ExpiresAt,
		"access_uuid": td.AccessUUID,
		"user_id":     publicID,
		"name":        username,
		"role":        role,
		"authorized":  true,
	})

	td.AccessToken, err = at.SignedString(a.secret)
	if err != nil {
		return nil, appErr.WrapWithCode(err, appErr.CodeHTTPInternalServerError, "Failed to sign access token")
	}

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":          td.ExpiresRt,
		"refresh_uuid": td.RefreshUUID,
		"user_id":      publicID,
		"name":         username,
		"role":         role,
	})

	td.RefreshToken, err = rt.SignedString(a.secret)
	if err != nil {
		return nil, appErr.WrapWithCode(err, appErr.CodeHTTPInternalServerError, "Failed to sign refresh token")
	}

	err = a.saveToRedis(ctx, publicID, td)
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (a *token) saveToRedis(ctx context.Context, publicID string, td *TokenDetails) error {
	ctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()

	respAccess := a.redis.Set(ctx, td.AccessUUID, publicID, a.expiredToken)
	if respAccess.Err() != nil {
		return appErr.WrapWithCode(respAccess.Err(), appErr.CodeHTTPInternalServerError, "failed to store access token in redis")
	}

	respRefresh := a.redis.Set(ctx, td.RefreshUUID, publicID, a.expiredRefreshToken)
	if respRefresh.Err() != nil {
		return appErr.WrapWithCode(respRefresh.Err(), appErr.CodeHTTPInternalServerError, "failed to store refresh token in redis")
	}

	return nil
}

func (a *token) ValidateToken(r *http.Request) (*AccessDetails, error) {
	return a.checkingToken(r)
}

func (a *token) checkingToken(r *http.Request) (*AccessDetails, error) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()

	tokenStr := a.extractToken(r)
	token, err := a.verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, preference.ErrInvalidToken)
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid user_id in token")
	}

	username, ok := claims["name"].(string)
	if !ok || username == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid name in token")
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid role in token")
	}

	accessUUID, ok := claims["access_uuid"].(string)
	if !ok || accessUUID == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid access_uuid in token")
	}

	redisIDUser, err := a.redis.Get(ctx, accessUUID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "access token expired")
		}

		return nil, appErr.WrapWithCode(err, appErr.CodeHTTPInternalServerError, "failed to get token from redis")
	}

	if userID != redisIDUser {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "authentication failure")
	}

	return &AccessDetails{
		AccessUUID: accessUUID,
		UserID:     redisIDUser,
		Username:   username,
		Role:       role,
	}, nil
}

func (a *token) extractToken(r *http.Request) string {
	authHeaders := r.Header["Authorization"]
	if len(authHeaders) == 0 {
		return ""
	}

	bearToken := authHeaders[0]
	if bearToken == "" {
		return ""
	}

	tokenArr := strings.Split(bearToken, " ")
	if len(tokenArr) == 2 {
		return tokenArr[1]
	}

	return ""
}

func (a *token) verifyToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(jwtToken *jwt.Token) (any, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, appErr.WrapWithCode(fmt.Errorf("unexpected signing method: %v", jwtToken.Header["alg"]), appErr.CodeHTTPUnauthorized, preference.ErrInvalidToken)
		}

		return a.secret, nil
	})
	if err != nil {
		return nil, appErr.WrapWithCode(err, appErr.CodeHTTPUnauthorized, preference.ErrInvalidToken)
	}

	return token, nil
}

func (a *token) ValidateRefreshToken(r *http.Request, tokenStr string) (*AccessDetails, error) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()

	token, err := a.verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, preference.ErrInvalidToken)
	}

	userID := a.getStringClaim(claims, "user_id")
	if userID == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid user_id in token")
	}

	username := a.getStringClaim(claims, "name")
	if username == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid name in token")
	}

	role := a.getStringClaim(claims, "role")
	if role == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid role in token")
	}

	refreshUUID := a.getStringClaim(claims, "refresh_uuid")
	if refreshUUID == "" {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "invalid refresh_uuid in token")
	}

	usedKey := RefreshTokenUsedPrefix + refreshUUID
	used, err := a.redis.Exists(ctx, usedKey).Result()
	if err == nil && used == 1 {
		a.redis.Del(ctx, a.getAccessTokenKey(userID))
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "refresh token has been used")
	}

	redisIDUser, err := a.redis.Get(ctx, refreshUUID).Result()
	if err != nil {
		return nil, appErr.WrapWithCode(err, appErr.CodeHTTPUnauthorized, "refresh token not found or expired")
	}

	if userID != redisIDUser {
		return nil, appErr.NewWithCode(appErr.CodeHTTPUnauthorized, "authentication failure")
	}

	if refreshTokenRotation {
		a.redis.Set(ctx, usedKey, "1", a.expiredRefreshToken)
		a.redis.Del(ctx, refreshUUID)
	}

	return &AccessDetails{
		RefreshUUID: refreshUUID,
		UserID:      userID,
		Username:    username,
		Role:        role,
	}, nil
}

func (a *token) getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}

	return ""
}

func (a *token) getAccessTokenKey(userID string) string {
	return preference.TokenKeyPrefix + userID
}
