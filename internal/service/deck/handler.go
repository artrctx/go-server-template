package deck

import (
	"context"
	"net/http"
	"uuid"

	"github.com/artrctx/templ/internal/auth"
	"github.com/artrctx/templ/internal/database/repository/deck"
	"github.com/artrctx/templ/internal/lib/httpu"
	"github.com/artrctx/templ/internal/lib/inst"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"
)

type ListPermittedRequestQuery struct {
	Limit       int32   `schema:"limit,default:10" validate:"gt=1"`
	Offset      int32   `schema:"offset,default:0" validate:"gte=0"`
	Name        *string `schema:"name" validate:"omitempty"`
	Description *string `schema:"description" validate:"omitempty"`
	Archived    *bool   `schema:"archived" validate:"omitempty"`
	Visibility  *string `schema:"visibility" validate:"omitempty,oneof=private public"`
}

type ListResponse struct {
	Decks []deck.Deck `json:"decks"`
	Total uint        `json:"total"`
}

func (d *Deck) ListPermitted(w http.ResponseWriter, r *http.Request) {
	usr, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httpu.Unauthorized(w)
		return
	}

	var query ListPermittedRequestQuery
	if err := inst.SchemaDecoder.Decode(&query, r.URL.Query()); err != nil {
		httpu.BadRequest(w, err)
		return
	}

	if err := inst.Validate.Struct(query); err != nil {
		httpu.BadRequest(w, err)
		return
	}

	eg, ctx := errgroup.WithContext(r.Context())
	var decks []deck.Deck
	var count int64
	eg.Go(func() (err error) {
		decks, err = d.deck.ListPermitted(ctx, deck.ListPermittedParams{UserID: usr.ID, Limit: query.Limit, Offset: query.Offset, Name: query.Name, Description: query.Description, Archived: query.Archived, Visibility: query.Visibility})
		return
	})
	eg.Go(func() (err error) {
		count, err = d.deck.PermittedCount(ctx, deck.PermittedCountParams{
			UserID:      usr.ID,
			Name:        query.Name,
			Description: query.Description, Archived: query.Archived, Visibility: query.Visibility,
		})
		return
	})

	if err := eg.Wait(); err != nil {
		httpu.InternalServerError(w, err)
		return
	}

	httpu.JsonEncode(w, r, http.StatusOK, ListResponse{decks, uint(count)})
}

type CreateRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1024"`
	Thumbnail   *string `json:"thumbnail" validate:"omitempty"`
	Visibility  *string `json:"visibility" validate:"omitempty,oneof=private public"`
}

func (d *Deck) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr, err := auth.UserFromCtx(ctx)
	if err != nil {
		httpu.Unauthorized(w)
		return
	}

	tx, err := d.db.Begin(ctx)
	if err != nil {
		httpu.InternalServerError(w, err)
		return
	}
	query := deck.New(tx)
	defer tx.Rollback(context.WithoutCancel(ctx))

	var req CreateRequest
	if err := httpu.JsonDecode(r.Body, &req); err != nil {
		httpu.BadRequest(w, err)
		return
	}

	if err := inst.Validate.Struct(req); err != nil {
		httpu.BadRequest(w, err)
		return
	}

	var visibility deck.DeckVisibility
	if req.Visibility != nil || *req.Visibility == "public" {
		visibility = deck.DeckVisibilityPublic
	} else {
		visibility = deck.DeckVisibilityPrivate
	}

	new, err := query.Create(ctx, deck.CreateParams{Name: req.Name, Description: req.Description, Thumbnail: req.Thumbnail, Visibility: visibility})

	if err != nil {
		// if future implementation have conflict update this
		httpu.InternalServerError(w, err)
		return
	}

	if err := query.LinkUser(ctx, deck.LinkUserParams{DeckID: new.ID, UserID: usr.ID}); err != nil {
		// if user can be linked in the future update tihs
		httpu.InternalServerError(w, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		httpu.InternalServerError(w, err)
		return
	}

	httpu.JsonEncode(w, r, http.StatusOK, new)
}

func (d *Deck) Find(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr, err := auth.UserFromCtx(ctx)
	if err != nil {
		httpu.Unauthorized(w)
		return
	}

	deckID, err := uuid.Parse(chi.URLParam(r, "deckID"))
	if err != nil {
		httpu.BadRequest(w, err)
		return
	}

	deck, err := d.deck.FindPermitted(ctx, deck.FindPermittedParams{ID: deckID, UserID: usr.ID})
	if err != nil {
		httpu.InternalServerError(w, err)
		return
	}

	httpu.JsonEncode(w, r, http.StatusOK, deck)
}

type UpdateRequest struct {
	Name        *string `json:"name" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1024"`
	Thumbnail   *string `json:"thumbnail" validate:"omitempty"`
	Visibility  *string `json:"visibility" validate:"omitempty,oneof=private public"`
}

func (d *Deck) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr, err := auth.UserFromCtx(ctx)
	if err != nil {
		httpu.Unauthorized(w)
		return
	}

	deckID, err := uuid.Parse(chi.URLParam(r, "deckID"))
	if err != nil {
		httpu.BadRequest(w, err)
		return
	}

	var body UpdateRequest
	if err := httpu.JsonDecode(r.Body, &body); err != nil {
		httpu.BadRequest(w, err)
		return
	}

	if err := inst.Validate.Struct(body); err != nil {
		httpu.BadRequest(w, err)
		return
	}

	var visibility deck.NullDeckVisibility
	if body.Visibility != nil {
		if *body.Visibility == "public" {
			visibility = deck.NullDeckVisibility{Valid: true, DeckVisibility: deck.DeckVisibilityPublic}
		} else {
			visibility = deck.NullDeckVisibility{Valid: true, DeckVisibility: deck.DeckVisibilityPrivate}
		}
	}

	updated, err := d.deck.Update(ctx, deck.UpdateParams{
		ID:          deckID,
		UserID:      usr.ID,
		Name:        body.Name,
		Description: body.Description,
		Thumbnail:   body.Thumbnail,
		Visibility:  visibility,
	})

	if err != nil {
		httpu.InternalServerError(w, err)
		return
	}

	httpu.JsonEncode(w, r, http.StatusOK, updated)
}
