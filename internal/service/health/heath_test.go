package health

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artrctx/shuffle-core/internal/database"
	"github.com/artrctx/shuffle-core/tests/dbtest"
)

var testDBConnStr string

func TestMain(m *testing.M) {
	dbtest.RunWithPostgres(m, &testDBConnStr)
}

func TestHealthHandler(t *testing.T) {
	dbServ, err := database.New(testDBConnStr)
	if err != nil {
		t.Fatalf("error creating db connection. err: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(HealthHandlerFunc(dbServ)))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("error making request to server. err: %v", err)
	}
	defer resp.Body.Close()

	// assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected Status OK; got %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading reasponse body. Err: %v", err)
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	if res["status"] != "healthy" {
		t.Errorf("expected response body status to be healthy; got %v", res["status"])
	}
}
