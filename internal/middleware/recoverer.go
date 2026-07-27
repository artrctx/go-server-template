package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/artrctx/shuffle-core/internal/lib/logger"
	"github.com/artrctx/shuffle-core/internal/lib/res"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(r)
				}

				logger.FromCtx(r.Context()).Error("Unknown Internal Server Error", slog.Any("error", rvr))

				if r.Header.Get("Connection") == "Upgrade" {
					return
				}
				res.InternalServerError(w, fmt.Errorf("Unknown Internal Server Error, %v", r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
