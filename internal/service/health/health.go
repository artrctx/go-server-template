package health

import (
	"fmt"
	"net/http"

	"github.com/artrctx/shuffle-core/internal/database"
	"github.com/artrctx/shuffle-core/internal/env"
	"github.com/artrctx/shuffle-core/internal/lib/json"
)

type Health struct {
	db *database.Service
}

type AuthHealthResp struct {
	Status string `json:"status"`
}

func authServerHealthStatus() string {
	resp, err := http.Get(fmt.Sprintf("%s/health", env.AuthUrl))
	if err != nil {
		return "down"
	}
	defer resp.Body.Close()
	var authRes AuthHealthResp
	if err := json.Decode(resp.Body, &authRes); err != nil {
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
			"service": "shuffle-server",
			"auth":    authServerHealthStatus(),
			"db":      dbStatus,
		}

		json.Encode(w, r, resStatus, response)
	}
}
