// Package audit records what an administrator did, which is what the activity
// list and the logs page are made of.
package audit

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"loginer/internal/api/session"
	"loginer/internal/model"
	"loginer/internal/store"
)

// Recorder writes those lines. Handlers that change something hold one.
type Recorder struct {
	store *store.Store
	log   *slog.Logger
}

// New returns a Recorder.
func New(st *store.Store, log *slog.Logger) Recorder {
	return Recorder{store: st, log: log}
}

// Record notes that the signed-in administrator did something to something.
//
// Failing to write the log must not fail the request that caused it: the
// change has already happened, so the error is logged and no more.
func (r Recorder) Record(c *gin.Context, action, targetType, targetID string) {
	r.RecordWith(c, action, targetType, targetID, nil)
}

// RecordWith is Record with more to keep about what happened: what else it
// involved, so the record can be found from there too. Never put secrets in
// it.
func (r Recorder) RecordWith(c *gin.Context, action, targetType, targetID string, metadata map[string]any) {
	actor := session.Admin(c)
	if actor == nil {
		return
	}

	id := actor.ID
	entry := model.AuditLog{
		AdminUserID: &id,
		ActorEmail:  actor.Username,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Metadata:    metadata,
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}

	if err := r.store.WriteAudit(c.Request.Context(), &entry); err != nil {
		r.log.Error("writing the activity log failed", "error", err, "action", action)
	}
}
