package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/artrctx/templ/internal/auth"
	"github.com/artrctx/templ/internal/lib/httpu"
	"github.com/artrctx/templ/internal/middleware"
	"github.com/artrctx/templ/internal/service/asset"
	"github.com/artrctx/templ/internal/service/deck"
	"github.com/artrctx/templ/internal/service/health"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func (s *Server) Register() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(chiMiddleware.ClientIPFromRemoteAddr)
	// 100 req per min rate limit
	r.Use(httprate.LimitBy(100, time.Minute, func(r *http.Request) (string, error) {
		return httprate.CanonicalizeIP(chiMiddleware.GetClientIP(r.Context())), nil
	}))

	r.Use(chiMiddleware.RequestID)

	var origins []string
	if s.env == "production" {
		origins = []string{`templ:\/\/*`}
	} else {
		origins = []string{`templ:\/\/*`, `exp:\/\/*`, `https:\/\/*`, `http:\/\/*`}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpu.NotFound(w, fmt.Errorf("invalid request path"))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpu.MethodNotAllowed(w)
	})

	// health
	r.Get("/health", health.HealthHandlerFunc(s.db))

	// auth
	r.Get("/auth/verify", auth.VerifyAuthHandler)

	// protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.ProtectedMiddleware)

		asset := asset.New(s.db, s.storage, s.bucket)
		r.Get("/assets/{entity}/{entityID}/*", asset.SteamAsset)
		r.Post("/assets/{entity}/{entityID}", asset.GenerateUploadAssetURL)

		deck := deck.New(s.db)
		r.Get("/me/decks", deck.ListPermitted)
		r.Get("/me/decks/{deckID}", deck.Find)
		r.Post("/decks", deck.Create)
		r.Patch("/decks/{deckID}", deck.Update)
	})

	return r
}
