package health

import (
	"fmt"
	"net/http"

	"github.com/artrctx/templ/internal/auth"
	"github.com/artrctx/templ/internal/database"
	"github.com/artrctx/templ/internal/lib/httpu"
)

type AuthHealthResp struct {
	Status string `json:"status"`
}

func authServerHealthStatus() string {
	resp, err := http.Get(fmt.Sprintf("%s/health", auth.BaseURL))
	if err != nil {
		return "down"
	}
	defer resp.Body.Close()
	var authRes AuthHealthResp
	if err := httpu.JsonDecode(resp.Body, &authRes); err != nil {
		return "down"
	}

	return authRes.Status
}

func HealthHandlerFunc(s *database.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "healthy"
		resStatus := http.StatusOK
		dbStatus := s.Health(r.Context())
		if dbStatus["status"] != "up" {
			resStatus = http.StatusInternalServerError
			status = "down"
		}

		response := map[string]interface{}{
			"status":  status,
			"service": "templ-server",
			"auth":    authServerHealthStatus(),
			"db":      dbStatus,
		}

		httpu.JsonEncode(w, r, resStatus, response)
	}
}
