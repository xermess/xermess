// Package database answers the endpoints behind the panel's Database page:
// the tables this server keeps its records in, and a page of any one of them.
//
// It reads and nothing else. There is no endpoint here that writes a row, and
// there is not meant to be: a user, an application or a role is edited on the
// page that knows what one is and what changing it costs. This is for looking
// — at what a migration actually built, and at the rows behind a page that is
// not showing what you expected.
//
// What it will not show is a password, a key, a one-time code or the hash
// standing in for a token. Those columns are never read (internal/store).
package database

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/respond"
	"xermess/internal/store"
)

// Handler holds what these endpoints need.
type Handler struct {
	store *store.Store
	log   *slog.Logger
}

// New returns a Handler.
func New(st *store.Store, log *slog.Logger) *Handler {
	return &Handler{store: st, log: log}
}

// Tables lists every table, with how many rows and columns each has.
func (h *Handler) Tables(c *gin.Context) {
	tables, err := h.store.Tables(c.Request.Context())
	if err != nil {
		respond.Failure(c, h.log, err, "listing the database tables failed")
		return
	}

	c.JSON(http.StatusOK, newTablesResponse(tables))
}

// Table returns one table's columns and a page of its rows.
func (h *Handler) Table(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("table")

	columns, err := h.store.TableColumns(ctx, name)
	switch {
	case errors.Is(err, store.ErrNotFound):
		respond.NotFound(c, "no such table")
		return
	case err != nil:
		respond.Failure(c, h.log, err, "describing a database table failed")
		return
	}

	var query pageRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		respond.BadRequest(c, "the query is not valid")
		return
	}

	limit, offset := query.page()

	rows, total, err := h.store.TableRows(ctx, name, limit, offset)
	if err != nil {
		respond.Failure(c, h.log, err, "reading a database table failed")
		return
	}

	c.JSON(http.StatusOK, newTableResponse(name, columns, rows, total, limit, offset))
}
