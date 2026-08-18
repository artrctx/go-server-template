package deck

import "github.com/artrctx/templ-core/internal/database"

type Deck struct {
	db *database.Service
}

func New(db *database.Service) Deck {
	return Deck{db}
}
