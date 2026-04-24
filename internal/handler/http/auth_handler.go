package rest

import (
	"encoding/json"
	"net/http"

	"go-far/internal/infra/validator"
	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
	appErr "go-far/internal/model/errors"

	"github.com/rs/zerolog"
)

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Create a new user account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"Registration data"
//	@Success		201		{object}	dto.HttpSuccessResp{data=entity.User}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		409		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/auth/register [post]
func (e *rest) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(w, r, appErr.WrapWithCode(err, appErr.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := validator.ValidateRequest(&req); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("validation_failed_register")
		e.httpRespError(w, r, err)
		return
	}

	user, err := e.usvc.RegisterUser(ctx, req)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusCreated, user, nil)
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user and return access/refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	dto.HttpSuccessResp{data=dto.TokenResponse}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		401		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/auth/login [post]
func (e *rest) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(w, r, appErr.WrapWithCode(err, appErr.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := validator.ValidateRequest(&req); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("validation_failed_login")
		e.httpRespError(w, r, err)
		return
	}

	user, err := e.usvc.Login(ctx, req)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	userData := userTokenData(user.ID, user.Name, user.Role)

	tokens, err := e.auth.GenerateToken(r, userData)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, dto.ToTokenResponse(tokens), nil)
}

// RefreshToken godoc
//
//	@Summary		Refresh access token
//	@Description	Generate a new access token using a valid refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshTokenRequest	true	"Refresh token"
//	@Success		200		{object}	dto.HttpSuccessResp{data=dto.TokenResponse}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		401		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/auth/refresh [post]
func (e *rest) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(w, r, appErr.WrapWithCode(err, appErr.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := validator.ValidateRequest(&req); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("validation_failed_refresh_token")
		e.httpRespError(w, r, err)
		return
	}

	accessDetails, err := e.auth.ValidateRefreshToken(r, req.RefreshToken)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	userData := userTokenData(accessDetails.UserID, accessDetails.Username, entity.ParseRole(accessDetails.Role))

	tokens, err := e.auth.GenerateToken(r, userData)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, dto.ToTokenResponse(tokens), nil)
}

// userTokenData creates a DTO for JWT token generation
func userTokenData(publicID, username string, role entity.Role) dto.UserTokenData {
	return dto.UserTokenData{
		PublicID: publicID,
		Username: username,
		Role:     role,
	}
}
