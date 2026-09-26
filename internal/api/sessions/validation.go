package sessions

import (
	"github.com/google/uuid"

	"loginer/internal/store"
)

// query turns the page's query string into what the store reads. The ids
// are already checked to be uuids, so they parse.
func (r listRequest) query(limit int) store.SessionQuery {
	q := store.SessionQuery{Search: r.Search, Limit: limit}

	if r.User != "" {
		q.UserID = uuid.MustParse(r.User)
	}
	if r.After != "" {
		q.After = uuid.MustParse(r.After)
	}

	return q
}
