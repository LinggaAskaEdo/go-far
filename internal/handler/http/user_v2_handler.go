package rest

import (
	"net/http"

	"go-far/internal/infra/middleware"
	"go-far/internal/model/dto"
	"go-far/internal/model/entity"
	x "go-far/internal/model/errors"
	"go-far/internal/util"
)

func (e *rest) ListUsersV2(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	filter := util.DecodeURL[dto.UserFilterV2](r.URL.Query())

	if authUser.Role != string(entity.RoleAdmin) {
		filter.ID = authUser.UserID
	}

	users, pagination, err := e.svc.User.ListUsersV2(ctx, &filter)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, users, pagination)
}
