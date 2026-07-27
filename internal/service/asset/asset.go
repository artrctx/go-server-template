package asset

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/artrctx/shuffle-core/internal/auth"
	"github.com/artrctx/shuffle-core/internal/database"
	"github.com/artrctx/shuffle-core/internal/database/repository/deck"
	"github.com/artrctx/shuffle-core/internal/env"
	"github.com/artrctx/shuffle-core/internal/lib/json"
	"github.com/artrctx/shuffle-core/internal/lib/res"
	"github.com/artrctx/shuffle-core/internal/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

type Asset struct {
	db      *database.Service
	storage *storage.Service
	sfGroup singleflight.Group
}

func New(db *database.Service, storage *storage.Service) Asset {
	return Asset{db, storage, singleflight.Group{}}
}

func (a *Asset) SteamAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr, err := auth.UserFromCtx(ctx)
	if err != nil {
		res.InternalServerError(w, err)
		return
	}
	entity, entityID, key := chi.URLParam(r, "entity"), chi.URLParam(r, "entityID"), chi.URLParam(r, "*")
	if entity == "" || entityID == "" || key == "" {
		res.BadRequest(w, fmt.Errorf("Entity, EntityID, and key is required ( /{entity}/{entityID}/{key} )"))
		return
	}

	switch entity {
	case "deck":
		deckID, err := uuid.Parse(entityID)
		if err != nil {
			res.BadRequest(w, err)
			return
		}
		//? might just want to cache result
		readable, err, _ := a.sfGroup.Do(fmt.Sprintf("deck-%s-%s", deckID, usr.ID), func() (any, error) {
			return a.db.Deck.CheckReadPermission(ctx, deck.CheckReadPermissionParams{UserID: usr.ID, DeckID: deckID})
		})
		if err != nil {
			res.InternalServerError(w, err)
			return
		}
		if val, _ := readable.(bool); val {
			break
		}

		res.NotFound(w, fmt.Errorf("deck asset not found"))
		return
	case "user":
		if entityID == usr.ID {
			break
		}
		res.BadRequest(w, fmt.Errorf("user asset not found"))
		return
	default:
		res.BadRequest(w, fmt.Errorf("%s is not a valid entity", entity))
		return
	}

	obj, err := a.storage.GetObject(ctx, env.R2BucketName, fmt.Sprintf("%s/%s/%s", entity, entityID, key))
	if err != nil {
		res.InternalServerError(w, err)
		return
	}
	defer obj.Body.Close()

	w.Header().Set("Content-Type", aws.ToString(obj.ContentType))
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("ETag", aws.ToString(obj.ETag))
	if obj.ContentLength != nil {
		w.Header().Set("Content-Length", strconv.FormatInt(*obj.ContentLength, 10))
	}
	io.Copy(w, obj.Body)
}

type GenerateUploadAssetURLRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
type GenerateUploadAssetURLResponse struct {
	URL      string      `json:"url"`
	Headers  http.Header `json:"headers"`
	Entity   string      `json:"entity"`
	EntityID string      `json:"entityID"`
	Key      string      `json:"key"`
}

func (a *Asset) GenerateUploadAssetURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr, err := auth.UserFromCtx(ctx)
	if err != nil {
		res.InternalServerError(w, err)
		return
	}

	var body GenerateUploadAssetURLRequest
	entity, entityID := chi.URLParam(r, "entity"), chi.URLParam(r, "entityID")
	if err := json.Decode(r.Body, &body); err != nil {
		res.BadRequest(w, err)
		return
	}

	switch entity {
	case "deck":
		deckID, err := uuid.Parse(entityID)
		if err != nil {
			res.BadRequest(w, err)
			return
		}

		allowed, err := a.db.Deck.CheckWritePermission(ctx, deck.CheckWritePermissionParams{UserID: usr.ID, DeckID: deckID})
		if err != nil {
			res.InternalServerError(w, err)
			return
		}
		if allowed {
			break
		}

		res.NotFound(w, fmt.Errorf("deck asset write not allowed"))
		return
	case "user":
		if entityID == usr.ID {
			break
		}
		res.BadRequest(w, fmt.Errorf("user asset write not found"))
		return
	default:
		res.BadRequest(w, fmt.Errorf("%s is not a valid entity type", entity))
		return
	}

	var kb strings.Builder
	if body.Path != "" {
		strings.TrimPrefix(body.Path, "/")
		kb.WriteString(body.Path)
		kb.WriteString("/")
	}
	if body.Name != "" {
		kb.WriteString(fmt.Sprintf("%s__%s", uuid.NewString(), body.Name))
	} else {
		kb.WriteString(uuid.NewString())
	}

	key := kb.String()
	presigned, err := a.storage.GeneratePresignedPut(ctx, storage.GeneratePresignedPutOpt{
		Bucket:  env.R2BucketName,
		Key:     fmt.Sprintf("%s/%s/%s", entity, entityID, key),
		Expires: 15 * time.Minute,
	})

	if err != nil {
		res.InternalServerError(w, err)
		return
	}

	json.Encode(w, r, http.StatusOK, GenerateUploadAssetURLResponse{
		URL:     presigned.URL,
		Headers: presigned.SignedHeader,
		Key:     key,
	})
}
