package rest

import (
	"net/http"

	"go-far/src/dto"
	x "go-far/src/errors"
	"go-far/src/preference"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// CreateUser godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user with the provided information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		dto.CreateUserRequest	true	"User data"
//	@Success		201		{object}	dto.HttpSuccessResp{data=domain.User}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/users [post]
func (e *rest) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	user, err := e.svc.User.CreateUser(ctx, req)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusCreated, user, nil)
}

// GetUser godoc
//
//	@Summary		Get user by ID
//	@Description	Get a user by their ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	dto.HttpSuccessResp{data=domain.User}
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/users/{id} [get]
func (e *rest) GetUser(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_user_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_user_id"))
		return
	}

	user, err := e.svc.User.GetUser(ctx, id.String())
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, user, nil)
}

// ListUsers godoc
//
//	@Summary		List users
//	@Description	Get a paginated list of users with optional filters
//	@Tags			users
//	@Produce		json
//	@Param			Cache-Control	header		string	false	"Request cache control"	Enums(must-revalidate, must-db-revalidate)
//	@Param			name			query		string	false	"Filter by name"
//	@Param			email			query		string	false	"Filter by email"
//	@Param			min_age			query		int		false	"Minimum age"
//	@Param			max_age			query		int		false	"Maximum age"
//	@Param			page			query		int		false	"Page number"	default(1)
//	@Param			page_size		query		int		false	"Page size"		default(10)
//	@Param			sort_by			query		string	false	"Sort by field"
//	@Param			sort_dir		query		string	false	"Sort direction (asc/desc)"	default(asc)
//	@Success		200				{object}	dto.HttpSuccessResp{data=[]domain.User}
//	@Failure		400				{object}	dto.HTTPErrorResp
//	@Failure		500				{object}	dto.HTTPErrorResp
//	@Router			/users [get]
func (e *rest) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()

	var (
		filter       dto.UserFilter
		cacheControl dto.CacheControl
	)

	if err := c.ShouldBindQuery(&filter); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_query_parameters")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_query_parameters"))
		return
	}

	if c.Request.Header[http.CanonicalHeaderKey(preference.CacheControl)] != nil && c.Request.Header[http.CanonicalHeaderKey(preference.CacheControl)][0] == preference.CacheMustRevalidate {
		cacheControl.MustRevalidate = true
	}

	users, pagination, err := e.svc.User.ListUsers(ctx, cacheControl, filter)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, users, &pagination)
}

// UpdateUser godoc
//
//	@Summary		Update user
//	@Description	Update an existing user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			user	body		dto.UpdateUserRequest	true	"User data"
//	@Success		200		{object}	dto.HttpSuccessResp{data=domain.User}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		404		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/users/{id} [put]
func (e *rest) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_user_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_user_id"))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	user, err := e.svc.User.UpdateUser(ctx, id.String(), req)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, user, nil)
}

// DeleteUser godoc
//
//	@Summary		Delete user
//	@Description	Delete a user by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	dto.HttpSuccessResp
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/users/{id} [delete]
func (e *rest) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_user_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_user_id"))
		return
	}

	if err := e.svc.User.DeleteUser(ctx, id.String()); err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, nil, nil)
}
