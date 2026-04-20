package rest

import (
	"net/http"

	"go-far/src/config/middleware"
	"go-far/src/model/dto"
	"go-far/src/model/entity"
	x "go-far/src/model/errors"
	"go-far/src/util"
)

func (e *rest) ListUsersV2(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := middleware.GetAuthUser(ctx)
	if !ok {
		e.httpRespError(w, r, x.NewWithCode(x.CodeHTTPUnauthorized, "unauthenticated"))
		return
	}

	filter := util.DecodeQuery[dto.UserFilterV2](r.URL.Query())

	if authUser.Role != string(entity.RoleAdmin) {
		filter.ID = authUser.UserID
	}

	users, pagination, err := e.svc.User.ListUsersV2(ctx, filter)
	if err != nil {
		e.httpRespError(w, r, err)
		return
	}

	e.httpRespSuccess(w, r, http.StatusOK, users, pagination)
}
