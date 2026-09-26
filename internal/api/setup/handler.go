// Package setup answers the two endpoints a new installation needs: whether
// it has an administrator yet, and the one request that creates the first
// one.
//
// Both are open, because there is nobody to sign in as yet. What keeps that
// from being a way in is the store: creating an administrator is refused the
// moment one exists, so this is a door that closes behind the first person
// through it and can never be opened again.
package setup

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"loginer/internal/api/respond"
	"loginer/internal/model"
	"loginer/internal/store"
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

// Status says whether the panel still has to be set up. The panel asks before
// showing its sign-in page, and sends whoever is there to the setup form
// instead when the answer is yes.
func (h *Handler) Status(c *gin.Context) {
	exists, err := h.store.AdminsExist(c.Request.Context())
	if err != nil {
		respond.Failure(c, h.log, err, "checking for administrators failed")
		return
	}

	c.JSON(http.StatusOK, statusResponse{Required: !exists})
}

// Create makes the first administrator: a super admin, with the address and
// password whoever is setting the panel up chose.
func (h *Handler) Create(c *gin.Context) {
	var req setupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating the first administrator failed")
		return
	}

	admin := model.AdminUser{
		// The address is the account: it is what this person signs in with,
		// and one less thing to invent during setup.
		Username:  req.Email,
		Email:     req.Email,
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Status:    model.StatusActive,
	}
	if err := admin.SetPassword(req.Password); err != nil {
		respond.Failure(c, h.log, err, "hashing the password failed")
		return
	}

	err := h.store.CreateFirstAdmin(c.Request.Context(), &admin)
	switch {
	case errors.Is(err, store.ErrAdminExists):
		respond.Conflict(c, "this panel has already been set up")
		return
	case err != nil:
		respond.Failure(c, h.log, err, "creating the first administrator failed")
		return
	}

	h.log.Info("the first administrator was created", "email", admin.Email)

	c.JSON(http.StatusCreated, gin.H{"admin": newAdminResponse(&admin)})
}
