package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/artrctx/shuffle-core/internal/env"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/httprc/v3/tracesink"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type UserKey string

const UserCtxKey UserKey = "shuffle-user"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

var (
	iMu      sync.Mutex
	JwkCache *jwk.Cache
)

// errors
var (
	ErrMissingUserID = errors.New("missing user id")
)

func loadCachedKeyset() error {
	iMu.Lock()
	defer iMu.Unlock()
	if JwkCache != nil {
		return nil
	}
	ctx := context.Background()
	c, err := jwk.NewCache(ctx, httprc.NewClient(
		httprc.WithTraceSink(tracesink.NewSlog(slog.New(slog.NewJSONHandler(os.Stderr, nil)))),
	))

	if err != nil {
		return err
	}

	if err := c.Register(ctx, env.JwksUrl); err != nil {
		return err
	}

	if _, err := c.Refresh(ctx, env.JwksUrl); err != nil {
		return err
	}

	JwkCache = c
	return nil
}

// https://pkg.go.dev/github.com/lestrrat-go/jwx/v3
func UserFromRequest(r *http.Request) (User, error) {
	if JwkCache == nil {
		if err := loadCachedKeyset(); err != nil {
			return User{}, fmt.Errorf("jwk cache: %w", err)
		}
	}

	timedCtx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	keyset, err := JwkCache.Lookup(timedCtx, env.JwksUrl)
	if err != nil {
		return User{}, fmt.Errorf("retrieve jwks: %w", err)
	}

	// header "Authorization" is default
	token, err := jwt.ParseRequest(r,
		jwt.WithKeySet(keyset),
		jwt.WithHeaderKey("Authorization"),
		// jwt.WithCookieKey("shuffle.session_token"),
		// jwt.WithCookieKey("__Secure-shuffle.session_token"),
	)

	if err != nil {
		return User{}, fmt.Errorf("jwt parse request: %w", err)
	}

	userID, exists := token.Subject()
	if !exists {
		return User{}, ErrMissingUserID
	}

	var email, name, role string

	token.Get("email", &email)
	token.Get("name", &name)
	token.Get("role", &role)

	return User{
		ID:    userID,
		Email: email,
		Name:  name,
		Role:  role,
	}, nil
}

func UserFromCtx(ctx context.Context) (*User, error) {
	if val, ok := ctx.Value(UserCtxKey).(User); ok {
		return &val, nil
	}
	return nil, fmt.Errorf("Unauthorized")
}

func Close() error {
	if JwkCache == nil {
		return nil
	}
	return JwkCache.Shutdown(context.Background())
}
