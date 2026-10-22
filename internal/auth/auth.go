package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/httprc/v3/tracesink"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type UserKey string

const UserCtxKey UserKey = "templ-user"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

var (
	jwkCache *jwk.Cache
	jwksURL  string
	BaseURL  string
)

// errors
var (
	ErrMissingUserID = errors.New("missing user id")
)

func Initialize(baseURL string) error {
	if jwkCache != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := jwk.NewCache(ctx, httprc.NewClient(
		httprc.WithTraceSink(tracesink.NewSlog(slog.New(slog.NewJSONHandler(os.Stderr, nil)))),
	))

	if err != nil {
		return err
	}

	BaseURL = baseURL
	jwksURL = fmt.Sprintf("%s/api/auth/jwks", baseURL)
	if err := c.Register(ctx, jwksURL); err != nil {
		return err
	}

	if _, err := c.Refresh(ctx, jwksURL); err != nil {
		return err
	}

	jwkCache = c

	return nil
}

// https://pkg.go.dev/github.com/lestrrat-go/jwx/v3
func UserFromRequest(r *http.Request) (User, error) {
	timedCtx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	keyset, err := jwkCache.Lookup(timedCtx, jwksURL)
	if err != nil {
		return User{}, fmt.Errorf("retrieve jwks: %w", err)
	}

	// header "Authorization" is default
	token, err := jwt.ParseRequest(r,
		jwt.WithKeySet(keyset),
		jwt.WithHeaderKey("Authorization"),
		// jwt.WithCookieKey("templ.session_token"),
		// jwt.WithCookieKey("__Secure-templ.session_token"),
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
	if jwkCache == nil {
		return nil
	}
	return jwkCache.Shutdown(context.Background())
}
