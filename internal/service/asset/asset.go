package asset

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/artrctx/templ/internal/auth"
	"github.com/artrctx/templ/internal/database"
	"github.com/artrctx/templ/internal/database/repository/deck"
	"github.com/artrctx/templ/internal/lib/httpu"
	"github.com/artrctx/templ/internal/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/singleflight"
)

type Asset struct {
	db      *database.Service
	deck    *deck.Queries
	storage *storage.Service
	bucket  string
	sfGroup singleflight.Group
}

func New(db *database.Service, storage *storage.Service, bucket string) Asset {
	return Asset{db, deck.New(db.Conn()), storage, bucket, singleflight.Group{}}
}

func (a *Asset) SteamAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr, err := auth.UserFromCtx(ctx)
	if err != nil {
		httpu.InternalServerError(w, err)
		return
	}
	entity, entityID, key := chi.URLParam(r, "entity"), chi.URLParam(r, "entityID"), chi.URLParam(r, "*")
	if entity == "" || entityID == "" || key == "" {
		httpu.BadRequest(w, fmt.Errorf("Entity, EntityID, and key is required ( /{entity}/{entityID}/{key} )"))
		return
	}

	switch entity {
	case "deck":
		deckID, err := uuid.Parse(entityID)
		if err != nil {
			httpu.BadRequest(w, err)
			return
		}
		//? might just want to cache result
		readable, err, _ := a.sfGroup.Do(fmt.Sprintf("deck-%s-%s", deckID, usr.ID), func() (any, error) {
			return a.deck.CheckReadPermission(ctx, deck.CheckReadPermissionParams{UserID: usr.ID, DeckID: deckID})
		})
		if err != nil {
			httpu.InternalServerError(w, err)
			return
		}
		if val, _ := readable.(bool); val {
			break
		}

		httpu.NotFound(w, fmt.Errorf("deck asset not found"))
		return
	case "user":
		if entityID == usr.ID {
			break
		}
		httpu.BadRequest(w, fmt.Errorf("user asset not found"))
		return
	default:
		httpu.BadRequest(w, fmt.Errorf("%s is not a valid entity", entity))
		return
	}

	obj, err := a.storage.GetObject(ctx, a.bucket, fmt.Sprintf("%s/%s/%s", entity, entityID, key))
	if err != nil {
		httpu.InternalServerError(w, err)
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
		httpu.InternalServerError(w, err)
		return
	}

	var body GenerateUploadAssetURLRequest
	entity, entityID := chi.URLParam(r, "entity"), chi.URLParam(r, "entityID")
	if err := httpu.JsonDecode(r.Body, &body); err != nil {
		httpu.BadRequest(w, err)
		return
	}

	switch entity {
	case "deck":
		deckID, err := uuid.Parse(entityID)
		if err != nil {
			httpu.BadRequest(w, err)
			return
		}

		allowed, err := a.deck.CheckWritePermission(ctx, deck.CheckWritePermissionParams{UserID: usr.ID, DeckID: deckID})
		if err != nil {
			httpu.InternalServerError(w, err)
			return
		}
		if allowed {
			break
		}

		httpu.NotFound(w, fmt.Errorf("deck asset write not allowed"))
		return
	case "user":
		if entityID == usr.ID {
			break
		}
		httpu.BadRequest(w, fmt.Errorf("user asset write not found"))
		return
	default:
		httpu.BadRequest(w, fmt.Errorf("%s is not a valid entity type", entity))
		return
	}

	var kb strings.Builder
	if body.Path != "" {
		strings.TrimPrefix(body.Path, "/")
		kb.WriteString(body.Path)
		kb.WriteString("/")
	}
	if body.Name != "" {
		kb.WriteString(fmt.Sprintf("%s__%s", uuid.New().String(), body.Name))
	} else {
		kb.WriteString(uuid.New().String())
	}

	key := kb.String()
	presigned, err := a.storage.GeneratePresignedPut(ctx, storage.GeneratePresignedPutOpt{
		Bucket:  a.bucket,
		Key:     fmt.Sprintf("%s/%s/%s", entity, entityID, key),
		Expires: 15 * time.Minute,
	})

	if err != nil {
		httpu.InternalServerError(w, err)
		return
	}

	httpu.JsonEncode(w, r, http.StatusOK, GenerateUploadAssetURLResponse{
		URL:     presigned.URL,
		Headers: presigned.SignedHeader,
		Key:     key,
	})
}
