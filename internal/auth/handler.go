package auth

import (
	"log/slog"
	"net/http"

	"github.com/artrctx/templ/internal/lib/httpu"
	"github.com/artrctx/templ/internal/lib/logger"
)

type AuthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	User    *User  `json:"user"`
}

func VerifyAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := UserFromRequest(r)
	if err != nil {
		logger.FromCtx(r.Context()).Error("failed to get user from request", slog.Any("error", err))
		httpu.Unauthorized(w)
		return
	}

	httpu.JsonEncode(w, r, http.StatusOK, AuthResponse{
		Status:  "success",
		Message: "Token is valid",
		User:    &user,
	})
}
