package rest

import (
	"math"
	"net/http"
	"strconv"

	"go-far/src/domain"
	exception "go-far/src/errors"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// CreateUser godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user with the provided information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		domain.CreateUserRequest	true	"User data"
//	@Success		201		{object}	domain.Response{data=domain.User}
//	@Failure		400		{object}	domain.Response
//	@Failure		500		{object}	domain.Response
//	@Router			/users [post]
func (r *rest) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Invalid request body")
		_ = c.Error(exception.BadRequestError("Invalid request body: " + err.Error()))
		return
	}

	user, err := r.svc.User.CreateUser(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(201, domain.SuccessResponse(user))
}

// GetUser godoc
//
//	@Summary		Get user by ID
//	@Description	Get a user by their ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	domain.Response{data=domain.User}
//	@Failure		404	{object}	domain.Response
//	@Failure		500	{object}	domain.Response
//	@Router			/users/{id} [get]
func (r *rest) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(exception.BadRequestError("Invalid user ID"))
		return
	}

	user, err := r.svc.User.GetUser(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(200, domain.SuccessResponse(user))
}

// ListUsers godoc
//
//	@Summary		List users
//	@Description	Get a paginated list of users with optional filters
//	@Tags			users
//	@Produce		json
//	@Param			name		query		string	false	"Filter by name"
//	@Param			email		query		string	false	"Filter by email"
//	@Param			min_age		query		int		false	"Minimum age"
//	@Param			max_age		query		int		false	"Maximum age"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(10)
//	@Param			sort_by		query		string	false	"Sort by field"
//	@Param			sort_dir	query		string	false	"Sort direction (asc/desc)"	default(asc)
//	@Success		200			{object}	domain.PaginatedResponse{data=[]domain.User}
//	@Failure		400			{object}	domain.Response
//	@Failure		500			{object}	domain.Response
//	@Router			/users [get]
func (e *rest) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()
	zerolog.Ctx(ctx).Debug().Msg("ListUsers called with context")
	var filter domain.UserFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		e.httpRespError(c, exception.BadRequestError("Invalid query parameters: "+err.Error()))
		return
	}

	users, total, err := e.svc.User.ListUsers(c.Request.Context(), filter)
	if err != nil {
		e.httpRespError(c, exception.InternalServerError("Error when listing users: "+err.Error()))
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	response := domain.PaginatedSuccessResponse(users, domain.MetaData{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	})

	e.httpRespSuccess(c, http.StatusOK, response)
}

// UpdateUser godoc
//
//	@Summary		Update user
//	@Description	Update an existing user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"User ID"
//	@Param			user	body		domain.UpdateUserRequest	true	"User data"
//	@Success		200		{object}	domain.Response{data=domain.User}
//	@Failure		400		{object}	domain.Response
//	@Failure		404		{object}	domain.Response
//	@Failure		500		{object}	domain.Response
//	@Router			/users/{id} [put]
func (e *rest) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		e.httpRespError(c, exception.BadRequestError("Invalid user ID"))
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		e.httpRespError(c, exception.BadRequestError("Invalid request body: "+err.Error()))
		return
	}

	user, err := e.svc.User.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		e.httpRespError(c, exception.InternalServerError("Error when updating user: "+err.Error()))
		return
	}

	e.httpRespSuccess(c, http.StatusOK, domain.SuccessResponse(user))
}

// DeleteUser godoc
//
//	@Summary		Delete user
//	@Description	Delete a user by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	domain.Response
//	@Failure		404	{object}	domain.Response
//	@Failure		500	{object}	domain.Response
//	@Router			/users/{id} [delete]
func (e *rest) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		e.httpRespError(c, exception.BadRequestError("Invalid user ID"))
		return
	}

	if err := e.svc.User.DeleteUser(c.Request.Context(), id); err != nil {
		e.httpRespError(c, exception.InternalServerError("Error when deleting user: "+err.Error()))
		return
	}

	e.httpRespSuccess(c, http.StatusOK, domain.SuccessResponse(gin.H{"message": "User deleted successfully"}))
}
