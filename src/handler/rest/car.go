package rest

import (
	"net/http"

	"go-far/src/dto"
	x "go-far/src/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// CreateCar godoc
//
//	@Summary		Create a new car
//	@Description	Create a new car with the provided information
//	@Tags			cars
//	@Accept			json
//	@Produce		json
//	@Param			car	body		dto.CreateCarRequest	true	"Car data"
//	@Success		201	{object}	dto.HttpSuccessResp{data=domain.Car}
//	@Failure		400	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars [post]
func (e *rest) CreateCar(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.CreateCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	car, err := e.svc.Car.CreateCar(ctx, req)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusCreated, car, nil)
}

// CreateBulkCars godoc
//
//	@Summary		Create multiple cars
//	@Description	Create multiple cars for a user in a single request
//	@Tags			cars
//	@Accept			json
//	@Produce		json
//	@Param			cars	body		dto.BulkCreateCarsRequest	true	"Cars data"
//	@Success		201		{object}	dto.HttpSuccessResp{data=[]domain.Car}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/cars/bulk [post]
func (e *rest) CreateBulkCars(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.BulkCreateCarsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	cars, err := e.svc.Car.CreateBulkCars(ctx, req)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusCreated, cars, nil)
}

// GetCar godoc
//
//	@Summary		Get car by ID
//	@Description	Get a car by its ID
//	@Tags			cars
//	@Produce		json
//	@Param			id	path		string	true	"Car ID"
//	@Success		200	{object}	dto.HttpSuccessResp{data=domain.Car}
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars/{id} [get]
func (e *rest) GetCar(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	car, err := e.svc.Car.GetCar(ctx, id)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, car, nil)
}

// GetCarWithOwner godoc
//
//	@Summary		Get car with owner details
//	@Description	Get a car by its ID with owner information
//	@Tags			cars
//	@Produce		json
//	@Param			id	path		string	true	"Car ID"
//	@Success		200	{object}	dto.HttpSuccessResp{data=domain.CarWithOwner}
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars/{id}/owner [get]
func (e *rest) GetCarWithOwner(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	car, err := e.svc.Car.GetCarWithOwner(ctx, id)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, car, nil)
}

// ListCarsByUser godoc
//
//	@Summary		List cars by user
//	@Description	Get all cars owned by a specific user
//	@Tags			cars
//	@Produce		json
//	@Param			user_id	path		string	true	"User ID"
//	@Success		200		{object}	dto.HttpSuccessResp{data=[]domain.Car}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/cars/by-user/{user_id} [get]
func (e *rest) ListCarsByUser(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_user_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_user_id"))
		return
	}

	cars, err := e.svc.Car.ListCarsByUser(ctx, userID)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, cars, nil)
}

// CountCarsByUser godoc
//
//	@Summary		Count cars by user
//	@Description	Get the total number of cars owned by a specific user
//	@Tags			cars
//	@Produce		json
//	@Param			user_id	path		string	true	"User ID"
//	@Success		200		{object}	dto.HttpSuccessResp{data=int}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/cars/by-user/{user_id}/count [get]
func (e *rest) CountCarsByUser(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_user_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_user_id"))
		return
	}

	count, err := e.svc.Car.CountCarsByUser(ctx, userID)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, count, nil)
}

// UpdateCar godoc
//
//	@Summary		Update car
//	@Description	Update an existing car
//	@Tags			cars
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Car ID"
//	@Param			car	body		dto.UpdateCarRequest	true	"Car data"
//	@Success		200	{object}	dto.HttpSuccessResp{data=domain.Car}
//	@Failure		400	{object}	dto.HTTPErrorResp
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars/{id} [put]
func (e *rest) UpdateCar(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	var req dto.UpdateCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	car, err := e.svc.Car.UpdateCar(ctx, id, req)
	if err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, car, nil)
}

// DeleteCar godoc
//
//	@Summary		Delete car
//	@Description	Delete a car by ID
//	@Tags			cars
//	@Produce		json
//	@Param			id	path		string	true	"Car ID"
//	@Success		200	{object}	dto.HttpSuccessResp
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars/{id} [delete]
func (e *rest) DeleteCar(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	if err := e.svc.Car.DeleteCar(ctx, id); err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, nil, nil)
}

// TransferCarOwnership godoc
//
//	@Summary		Transfer car ownership
//	@Description	Transfer a car to a new owner
//	@Tags			cars
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Car ID"
//	@Param			request	body		dto.TransferCarRequest	true	"New owner ID"
//	@Success		200		{object}	dto.HttpSuccessResp
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		404		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/cars/{id}/transfer [post]
func (e *rest) TransferCarOwnership(c *gin.Context) {
	ctx := c.Request.Context()

	carID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	var req dto.TransferCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := e.svc.Car.TransferCarOwnership(ctx, carID, req.NewUserID); err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, nil, nil)
}

// BulkUpdateAvailability godoc
//
//	@Summary		Bulk update car availability
//	@Description	Update availability status for multiple cars
//	@Tags			cars
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.BulkUpdateAvailabilityRequest	true	"Car IDs and availability status"
//	@Success		200		{object}	dto.HttpSuccessResp
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/cars/availability [put]
func (e *rest) BulkUpdateAvailability(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.BulkUpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(c, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := e.svc.Car.BulkUpdateAvailability(ctx, req); err != nil {
		e.httpRespError(c, err)
		return
	}

	e.httpRespSuccess(c, http.StatusOK, nil, nil)
}
