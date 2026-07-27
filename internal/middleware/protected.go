package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/artrctx/shuffle-core/internal/auth"
	"github.com/artrctx/shuffle-core/internal/lib/logger"
	"github.com/artrctx/shuffle-core/internal/lib/res"
)

func Protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rootCtx := r.Context()
		user, err := auth.UserFromRequest(r)

		if err != nil {
			logger.FromCtx(rootCtx).Error("failed to get user from request", slog.Any("error", err))
			res.Unauthorized(w)
			return
		}

		ctx := logger.CtxWithLogger(rootCtx, logger.FromCtx(rootCtx).With(
			"user_id", user.ID,
			"user_name", user.Name,
		))

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, auth.UserCtxKey, user)))
	})
}
