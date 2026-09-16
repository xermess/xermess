// Package activity answers the endpoints that report on what has been
// happening: the dashboard's overview, and the log of what administrators did.
package activity

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/query"
	"xermess/internal/api/respond"
	"xermess/internal/api/session"
	"xermess/internal/model"
	"xermess/internal/store"
)

// How much the endpoints return.
const (
	// dashboardEntries is how many of the latest entries the dashboard lists.
	dashboardEntries = 12
	// chartDays is how many days the activity chart covers, today included.
	chartDays = 14
	// topActors is how many of the busiest administrators are named.
	topActors = 5
	// recentWindow is what "recently" means on the dashboard: new users,
	// sign-ins and the busiest administrators are counted over it.
	recentWindow = 7 * 24 * time.Hour

	defaultLimit = 50
	maxLimit     = 200
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

// Overview is the admin panel's front page: how much of everything there is,
// how signing in has gone, what each of the last days looked like, who has
// been busiest, and the latest entries.
func (h *Handler) Overview(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()
	since := now.Add(-recentWindow)

	counts, err := h.store.Counts(ctx, now, since)
	if err != nil {
		respond.Failure(c, h.log, err, "counting for the overview failed")
		return
	}

	signIns, err := h.store.SignInsSince(ctx, since)
	if err != nil {
		respond.Failure(c, h.log, err, "counting sign-ins failed")
		return
	}

	daily, err := h.store.DailyActivity(ctx, now, chartDays)
	if err != nil {
		respond.Failure(c, h.log, err, "counting daily activity failed")
		return
	}

	actors, err := h.store.TopActors(ctx, since, topActors)
	if err != nil {
		respond.Failure(c, h.log, err, "finding the busiest administrators failed")
		return
	}

	events, err := h.store.AuditLog(ctx, dashboardEntries)
	if err != nil {
		respond.Failure(c, h.log, err, "reading the activity log failed")
		return
	}

	activity, err := h.describe(c, events)
	if err != nil {
		respond.Failure(c, h.log, err, "naming what the activity was about failed")
		return
	}

	c.JSON(http.StatusOK, overviewResponse{
		Counts:    counts,
		SignIns:   signIns,
		Daily:     daily,
		TopActors: actors,
		Activity:  activity,
	})
}

// Logs lists the activity log, newest first. The page size is capped so a
// caller cannot ask for the whole table.
func (h *Handler) Logs(c *gin.Context) {
	events, err := h.store.AuditLog(c.Request.Context(), query.Int(c, "limit", defaultLimit, maxLimit))
	if err != nil {
		respond.Failure(c, h.log, err, "listing logs failed")
		return
	}

	described, err := h.describe(c, events)
	if err != nil {
		respond.Failure(c, h.log, err, "naming what the logs were about failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": newLogResponses(events, described)})
}

// describe turns log entries into what the panel shows: each with its target
// named, where the administrator may see that target, and the detail worth
// reading.
func (h *Handler) describe(c *gin.Context, events []model.AuditLog) ([]eventResponse, error) {
	ids := map[string][]string{}
	for _, event := range events {
		if event.TargetType != "" && event.TargetID != "" {
			ids[event.TargetType] = append(ids[event.TargetType], event.TargetID)
		}
	}

	names := map[string]map[string]store.TargetName{}
	for targetType, of := range ids {
		found, err := h.store.TargetNames(c.Request.Context(), targetType, of)
		if err != nil {
			return nil, err
		}
		names[targetType] = found
	}

	out := make([]eventResponse, 0, len(events))
	for _, event := range events {
		response := newEventResponse(event)

		if event.TargetType != "" {
			response.Target = &targetResponse{Type: event.TargetType, ID: event.TargetID}

			if name, ok := names[event.TargetType][event.TargetID]; ok && maySeeName(c, event.TargetType, name) {
				response.Target.Name = name.Name
			}
		}

		// What a user did at the sign-in pages is recorded under their own
		// address, which is only for administrators who may read users.
		if event.AdminUserID == nil && event.TargetType == "user" && !maySeeName(c, "user", store.TargetName{}) {
			response.Actor = "A user"
		}

		response.Detail = detail(c, event)
		out = append(out, response)
	}

	return out, nil
}

// maySeeName reports whether the administrator may be told what a target is
// called. Reading the log shows that something happened to a record; what
// the record is called is the business of whoever may read that kind of
// record.
func maySeeName(c *gin.Context, targetType string, name store.TargetName) bool {
	admin := session.Admin(c)
	if admin == nil {
		return false
	}

	switch targetType {
	case "user", "user_field":
		return admin.HasPermission(model.PermUsersRead)
	case "application":
		return name.ApplicationID != nil && session.Allowed(c, model.PermApplicationsRead, *name.ApplicationID)
	case "user_role":
		return name.ApplicationID == nil || session.Allowed(c, model.PermApplicationsRead, *name.ApplicationID)
	case "api":
		return admin.HasPermission(model.PermAPIsRead)
	case "admin_user", "admin_role":
		return admin.IsSuperAdmin()
	case "organization":
		return admin.HasPermission(model.PermOrganizationRead)
	default:
		return false
	}
}

// detail is what an entry's metadata adds that is worth a line: why a sign-in
// was refused, or which API an application gained or lost.
func detail(c *gin.Context, event model.AuditLog) string {
	text := func(key string) string {
		value, _ := event.Metadata[key].(string)
		return value
	}

	// A list comes back from the log as JSON did: whatever was stored, in
	// whatever type it decoded to.
	list := func(key string) string {
		values, _ := event.Metadata[key].([]any)

		parts := make([]string, 0, len(values))
		for _, value := range values {
			if part, ok := value.(string); ok {
				parts = append(parts, part)
			}
		}

		return strings.Join(parts, ", ")
	}

	switch event.Action {
	case "admin.login_failed", "admin.login_blocked", "admin.mfa_failed", "user.login_failed", "user.login_blocked":
		return text("reason")
	case "application.api_authorized", "application.api_revoked":
		if admin := session.Admin(c); admin != nil && admin.HasPermission(model.PermAPIsRead) {
			return text("api")
		}
	case "organization.updated":
		// Which settings moved, for whoever may see them at all.
		if admin := session.Admin(c); admin != nil && admin.HasPermission(model.PermOrganizationRead) {
			return list("fields")
		}
	}

	return ""
}
