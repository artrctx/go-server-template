package httpu

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/artrctx/templ/internal/lib/logger"
)

func JsonEncode(w http.ResponseWriter, r *http.Request, statusCode int, v any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		InternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(buf.Bytes()); err != nil {
		if r == nil {
			return
		}
		logger.FromCtx(r.Context()).Error("buffer write failed", slog.Any("error", err))
	}
}

func JsonDecode(r io.ReadCloser, v any) error {
	return json.NewDecoder(r).Decode(v)
}
