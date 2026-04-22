package rest

import (
	"encoding/json"
	"net/http"

	"go-far/src/config/middleware"
	"go-far/src/config/validator"
	"go-far/src/model/dto"
	"go-far/src/model/entity"
	x "go-far/src/model/errors"

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
//	@Success		201	{object}	dto.HttpSuccessResp{data=entity.Car}
//	@Failure		400	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars [post]
func (e *rest) CreateCar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	var req dto.CreateCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := validator.ValidateRequest(&req); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("validation_failed_create_car")
		e.httpRespError(w, r, err)
		return
	}

	car, err := e.svc.Car.CreateCar(ctx, req, authUser.UserID)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusCreated, car, nil)
}

// CreateBulkCars godoc
//
//	@Summary		Create multiple cars
//	@Description	Create multiple cars for a user in a single request
//	@Tags			cars
//	@Accept			json
//	@Produce		json
//	@Param			cars	body		dto.BulkCreateCarsRequest	true	"Cars data"
//	@Success		201		{object}	dto.HttpSuccessResp{data=[]entity.Car}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/cars/bulk [post]
func (e *rest) CreateBulkCars(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	var req dto.BulkCreateCarsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := validator.ValidateRequest(&req); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("validation_failed_create_bulk_cars")
		e.httpRespError(w, r, err)
		return
	}

	cars, err := e.svc.Car.CreateBulkCars(ctx, req, authUser.UserID)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusCreated, cars, nil)
}

// GetCar godoc
//
//	@Summary		Get car by ID
//	@Description	Get a car by its ID
//	@Tags			cars
//	@Produce		json
//	@Param			id	path		string	true	"Car ID"
//	@Success		200	{object}	dto.HttpSuccessResp{data=entity.Car}
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars/{id} [get]
func (e *rest) GetCar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	car, err := e.svc.Car.GetCar(ctx, id)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, car, nil)
}

// GetCarWithOwner godoc
//
//	@Summary		Get car with owner details
//	@Description	Get a car by its ID with owner information
//	@Tags			cars
//	@Produce		json
//	@Param			id	path		string	true	"Car ID"
//	@Success		200	{object}	dto.HttpSuccessResp{data=entity.CarWithOwner}
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars/{id}/owner [get]
func (e *rest) GetCarWithOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	car, err := e.svc.Car.GetCarWithOwner(ctx, id)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, car, nil)
}

// ListCarsByUser godoc
//
//	@Summary		List cars by user
//	@Description	Get all cars owned by a specific user
//	@Tags			cars
//	@Produce		json
//	@Param			user_id	path		string	true	"User ID"
//	@Success		200		{object}	dto.HttpSuccessResp{data=[]entity.Car}
//	@Failure		400		{object}	dto.HTTPErrorResp
//	@Failure		403		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/users/{user_id}/cars [get]
func (e *rest) ListCarsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_user_id")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_user_id"))
		return
	}

	// Non-admin users can only view their own cars
	if authUser.Role != string(entity.RoleAdmin) && authUser.UserID != userID.String() {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPForbidden, "forbidden"))
		return
	}

	cars, err := e.svc.Car.ListCarsByUser(ctx, userID)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, cars, nil)
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
//	@Failure		403		{object}	dto.HTTPErrorResp
//	@Failure		500		{object}	dto.HTTPErrorResp
//	@Router			/users/{user_id}/cars/count [get]
func (e *rest) CountCarsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_user_id")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_user_id"))
		return
	}

	// Non-admin users can only count their own cars
	if authUser.Role != string(entity.RoleAdmin) && authUser.UserID != userID.String() {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPForbidden, "forbidden"))
		return
	}

	count, err := e.svc.Car.CountCarsByUser(ctx, userID)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, count, nil)
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
//	@Success		200	{object}	dto.HttpSuccessResp{data=entity.Car}
//	@Failure		400	{object}	dto.HTTPErrorResp
//	@Failure		404	{object}	dto.HTTPErrorResp
//	@Failure		500	{object}	dto.HTTPErrorResp
//	@Router			/cars/{id} [put]
func (e *rest) UpdateCar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	var req dto.UpdateCarRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		zerolog.Ctx(ctx).Error().Err(decodeErr).Msg("invalid_request_body")
		e.httpRespError(w, r, x.WrapWithCode(decodeErr, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if validateErr := validator.ValidateRequest(&req); validateErr != nil {
		zerolog.Ctx(ctx).Warn().Err(validateErr).Msg("validation_failed_update_car")
		e.httpRespError(w, r, validateErr)
		return
	}

	car, err := e.svc.Car.UpdateCar(ctx, id, &req, authUser.UserID)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, car, nil)
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
func (e *rest) DeleteCar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	if err := e.svc.Car.DeleteCar(ctx, id, authUser.UserID); err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, nil, nil)
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
func (e *rest) TransferCarOwnership(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	carID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_car_id")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPBadRequest, "invalid_car_id"))
		return
	}

	var req dto.TransferCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := e.svc.Car.TransferCarOwnership(ctx, carID, req.NewUserID, authUser.UserID); err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, nil, nil)
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
func (e *rest) BulkUpdateAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	var req dto.BulkUpdateAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("invalid_request_body")
		e.httpRespError(w, r, x.WrapWithCode(err, x.CodeHTTPUnmarshal, "invalid_request_body"))
		return
	}

	if err := e.svc.Car.BulkUpdateAvailability(ctx, req, authUser.UserID); err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, nil, nil)
}
