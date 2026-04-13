package token

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	x "go-far/src/model/errors"
	"go-far/src/preference"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

const jwtSecretEnv = "JWT_SECRET_GO_FAR"

// Token defines the token management interface
type Token interface {
	GenerateToken(r *http.Request, data any) (*TokenDetails, error)
	ValidateToken(r *http.Request) (*AccessDetails, error)
	ValidateRefreshToken(r *http.Request, token string) (*AccessDetails, error)
}

var (
	onceToken = &sync.Once{}
	tokenInst *token
)

// TokenOptions holds token configuration
type TokenOptions struct {
	ExpiredToken        time.Duration `yaml:"expired_token"`
	ExpiredRefreshToken time.Duration `yaml:"expired_refresh_token"`
}

type token struct {
	log                 zerolog.Logger
	redis               *redis.Client
	secret              []byte
	expiredToken        time.Duration
	expiredRefreshToken time.Duration
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
func InitToken(log zerolog.Logger, opt TokenOptions, redis *redis.Client) Token {
	onceToken.Do(func() {
		secret := os.Getenv(jwtSecretEnv)
		if secret == "" {
			log.Panic().Msgf("Environment variable %s is not set", jwtSecretEnv)
		}

		tokenInst = &token{
			log:                 log,
			redis:               redis,
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
		return nil, x.NewWithCode(x.CodeHTTPBadRequest, "Invalid data for token generation")
	}

	publicIDField := dataVal.FieldByName("PublicID")
	usernameField := dataVal.FieldByName("Username")
	roleField := dataVal.FieldByName("Role")

	if !publicIDField.IsValid() || !usernameField.IsValid() || !roleField.IsValid() {
		return nil, x.NewWithCode(x.CodeHTTPBadRequest, "Data must contain PublicID, Username and Role fields")
	}

	publicID := publicIDField.String()
	username := usernameField.String()
	role := roleField.String()

	if publicID == "" || username == "" || role == "" {
		return nil, x.NewWithCode(x.CodeHTTPBadRequest, "PublicID, Username and Role cannot be empty")
	}

	td.ExpiresAt = time.Now().Add(a.expiredToken).Unix()
	td.AccessUUID = ksuid.New().String()

	td.ExpiresRt = time.Now().Add(a.expiredRefreshToken).Unix()
	td.RefreshUUID = td.AccessUUID + "++" + publicID

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
		return nil, x.WrapWithCode(err, x.CodeHTTPInternalServerError, "Failed to sign access token")
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
		return nil, x.WrapWithCode(err, x.CodeHTTPInternalServerError, "Failed to sign refresh token")
	}

	err = a.saveToRedis(ctx, publicID, td)
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (a *token) saveToRedis(ctx context.Context, publicID string, td *TokenDetails) error {
	respAccess := a.redis.Set(ctx, td.AccessUUID, publicID, a.expiredToken)
	if respAccess.Err() != nil {
		return x.WrapWithCode(respAccess.Err(), x.CodeHTTPInternalServerError, "Failed to store access token in Redis")
	}

	respRefresh := a.redis.Set(ctx, td.RefreshUUID, publicID, a.expiredRefreshToken)
	if respRefresh.Err() != nil {
		return x.WrapWithCode(respRefresh.Err(), x.CodeHTTPInternalServerError, "Failed to store refresh token in Redis")
	}

	return nil
}

func (a *token) ValidateToken(r *http.Request) (*AccessDetails, error) {
	return a.checkingToken(r)
}

func (a *token) checkingToken(r *http.Request) (*AccessDetails, error) {
	ctx := r.Context()

	tokenStr := a.extractToken(r)
	token, err := a.verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, x.NewWithCode(x.CodeHTTPUnauthorized, preference.ErrInvalidToken)
	}

	userID := claims["user_id"].(string)
	username := claims["name"].(string)
	role := claims["role"].(string)

	var accessUUID, redisIDUser string

	accessUUID, ok = claims["access_uuid"].(string)
	if !ok {
		return nil, x.NewWithCode(x.CodeHTTPUnauthorized, "Failed claims accessUUID")
	}

	redisIDUser, err = a.redis.Get(ctx, accessUUID).Result()
	if err != nil {
		return nil, x.WrapWithCode(err, x.CodeHTTPInternalServerError, "Failed to get token from Redis")
	}

	if userID != redisIDUser {
		return nil, x.NewWithCode(x.CodeHTTPUnauthorized, "Authentication failure")
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
	if len(bearToken) == 0 {
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
			return nil, x.WrapWithCode(fmt.Errorf("unexpected signing method: %v", jwtToken.Header["alg"]), x.CodeHTTPUnauthorized, preference.ErrInvalidToken)
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, x.WrapWithCode(err, x.CodeHTTPUnauthorized, preference.ErrInvalidToken)
	}

	return token, nil
}

func (a *token) ValidateRefreshToken(r *http.Request, tokenStr string) (*AccessDetails, error) {
	ctx := r.Context()

	token, err := a.verifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, x.NewWithCode(x.CodeHTTPUnauthorized, preference.ErrInvalidToken)
	}

	userID := claims["user_id"].(string)
	username := claims["name"].(string)
	role := claims["role"].(string)

	var accessUUID, refreshUUID, redisIDUser string

	refreshUUID, ok = claims["refresh_uuid"].(string)
	if !ok {
		return nil, x.NewWithCode(x.CodeHTTPUnauthorized, "Failed claims refresh_uuid")
	}

	redisIDUser, err = a.redis.Get(ctx, refreshUUID).Result()
	if err != nil {
		return nil, x.WrapWithCode(err, x.CodeHTTPInternalServerError, "Failed to get token from Redis")
	}

	if userID != redisIDUser {
		return nil, x.NewWithCode(x.CodeHTTPUnauthorized, "Authentication failure")
	}

	return &AccessDetails{
		AccessUUID:  accessUUID,
		RefreshUUID: refreshUUID,
		UserID:      redisIDUser,
		Username:    username,
		Role:        role,
	}, nil
}
