package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/artrctx/templ/internal/lib/httpu"
	"github.com/artrctx/templ/internal/lib/logger"
)

func ProtectedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rootCtx := r.Context()
		user, err := UserFromRequest(r)

		if err != nil {
			logger.FromCtx(rootCtx).Error("failed to get user from request", slog.Any("error", err))
			httpu.Unauthorized(w)
			return
		}

		ctx := logger.CtxWithLogger(rootCtx, logger.FromCtx(rootCtx).With(
			"user_id", user.ID,
			"user_name", user.Name,
		))

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, UserCtxKey, user)))
	})
}
