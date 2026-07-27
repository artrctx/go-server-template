package server

import (
	"fmt"
	"net/http"

	"github.com/artrctx/shuffle-core/internal/env"
	"github.com/artrctx/shuffle-core/internal/lib/res"
	"github.com/artrctx/shuffle-core/internal/middleware"
	"github.com/artrctx/shuffle-core/internal/service/asset"
	"github.com/artrctx/shuffle-core/internal/service/auth"
	"github.com/artrctx/shuffle-core/internal/service/deck"
	"github.com/artrctx/shuffle-core/internal/service/health"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) Register() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Recoverer)

	var origins []string
	if env.AppEnv == "production" {
		origins = []string{`shuffle:\/\/*`}
	} else {
		origins = []string{`shuffle:\/\/*`, `exp:\/\/*`, `https:\/\/*`, `http:\/\/*`}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		res.NotFound(w, fmt.Errorf("invalid request path"))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		res.MethodNotAllowed(w)
	})

	// health
	r.Get("/health", health.HealthHandlerFunc(s.db))

	// auth
	r.Get("/auth/verify", auth.VerifyAuthHandler)

	// protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Protected)

		asset := asset.New(s.db, s.storage)
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
