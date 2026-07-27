package auth

import (
	"log/slog"
	"net/http"

	"github.com/artrctx/shuffle-core/internal/auth"
	"github.com/artrctx/shuffle-core/internal/lib/json"
	"github.com/artrctx/shuffle-core/internal/lib/logger"
	"github.com/artrctx/shuffle-core/internal/lib/res"
)

type AuthResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	User    *auth.User `json:"user"`
}

func VerifyAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := auth.UserFromRequest(r)
	if err != nil {
		logger.FromCtx(r.Context()).Error("failed to get user from request", slog.Any("error", err))
		res.Unauthorized(w)
		return
	}

	json.Encode(w, r, http.StatusOK, AuthResponse{
		Status:  "success",
		Message: "Token is valid",
		User:    &user,
	})
}
