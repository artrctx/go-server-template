package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/artrctx/shuffle-core/internal/lib/logger"
	md "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type BodyLogWriter struct {
	md.WrapResponseWriter
	bodyBuf       *bytes.Buffer
	shouldLog     bool
	statusWritten bool
}

func (w *BodyLogWriter) WriteHeader(statusCode int) {
	w.WrapResponseWriter.WriteHeader(statusCode)
	if w.statusWritten {
		return
	}

	w.shouldLog = statusCode >= 400
	w.statusWritten = true
}

func (w *BodyLogWriter) Write(b []byte) (int, error) {
	// if status is not written defaults to 200
	if !w.statusWritten {
		w.shouldLog = false
		w.statusWritten = true
	}

	if w.shouldLog {
		w.bodyBuf.Write(b)
	}

	return w.WrapResponseWriter.Write(b)
}

func getClientIp(r *http.Request) string {
	// 1. Cloudflare's official visitor IP header (Always trust this with Tunnels)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// 2. Fallback for standard proxies (e.g., if you route through TrueClientIP)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}

	// 3. Local fallback
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// Need Idempotency key too
// https://medium.com/@quentinsims89/logging-in-go-structured-contextual-and-built-into-the-standard-library-e86027563108
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l := slog.Default().With(
			"request_id", uuid.New().String(),
			"method", r.Method,
			"path", r.URL.Path,
			"ip", getClientIp(r),
		)

		ctx := logger.CtxWithLogger(r.Context(), l)

		r = r.WithContext(ctx)
		blw := &BodyLogWriter{
			WrapResponseWriter: md.NewWrapResponseWriter(w, r.ProtoMajor),
			bodyBuf:            new(bytes.Buffer),
		}

		next.ServeHTTP(blw, r)

		statusCode := blw.Status()

		l = l.With(
			"status", statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)

		reqMsg := fmt.Sprintf("HTTP Request | %d | %s | %s", statusCode, r.Method, r.URL.Path)

		if statusCode < 400 {
			l.Info(reqMsg)
		} else {
			l.Error(reqMsg, slog.Any("error", blw.bodyBuf.String()))
		}
	})
}
