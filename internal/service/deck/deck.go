package deck

import (
	"github.com/artrctx/templ/internal/database"
	"github.com/artrctx/templ/internal/database/repository/deck"
)

type Deck struct {
	db   *database.Service
	deck *deck.Queries
}

func New(db *database.Service) Deck {
	return Deck{db, deck.New(db.Conn())}
}
