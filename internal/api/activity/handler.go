// Package activity answers the endpoints that report on what has been
// happening: the dashboard's overview, and the log of what administrators did.
package activity

import (
	"encoding/csv"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"loginer/internal/api/query"
	"loginer/internal/api/respond"
	"loginer/internal/api/session"
	"loginer/internal/model"
	"loginer/internal/store"
)

// How much the endpoints return.
const (
	// dashboardEntries is how many of the latest entries the dashboard lists.
	dashboardEntries = 12
	// topActors is how many of the busiest administrators are named.
	topActors = 5
	// defaultDays is the dashboard's range when it asks for none.
	defaultDays = 14

	defaultLimit = 50
	maxLimit     = 200

	// exportLimit caps an export, and exportPage is how many entries it
	// reads at a time: a file of the whole history is a job for the
	// database's own tools, not a request.
	exportLimit = 10000
	exportPage  = 500
)

// ranges are the spans the dashboard may be looked at over, in days. A
// short list rather than any number: each is a chart the page knows how to
// draw, and a year of daily bars is not one.
var ranges = map[int]bool{7: true, 14: true, 30: true, 90: true}

// Handler holds what these endpoints need.
type Handler struct {
	store *store.Store
	log   *slog.Logger
}

// New returns a Handler.
func New(st *store.Store, log *slog.Logger) *Handler {
	return &Handler{store: st, log: log}
}

// Overview is the admin panel's front page over a range of days: how much of
// everything there is, how signing in has gone, what each day looked like,
// who has been busiest, and the latest entries.
//
// The five reads do not depend on each other, so they run at once: the page
// waits for the slowest of them rather than for all of them in a row.
func (h *Handler) Overview(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()

	days := query.Int(c, "days", defaultDays, 90)
	if !ranges[days] {
		days = defaultDays
	}
	since := now.AddDate(0, 0, -days)

	var (
		counts  store.Counts
		signIns store.SignIns
		daily   []store.DayCount
		actors  []store.ActorCount
		events  []model.AuditLog
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() (err error) { counts, err = h.store.Counts(groupCtx, now, since); return })
	group.Go(func() (err error) { signIns, err = h.store.SignInsSince(groupCtx, since); return })
	group.Go(func() (err error) { daily, err = h.store.DailyActivity(groupCtx, now, days); return })
	group.Go(func() (err error) { actors, err = h.store.TopActors(groupCtx, since, topActors); return })
	group.Go(func() (err error) { events, err = h.store.AuditLog(groupCtx, dashboardEntries); return })

	if err := group.Wait(); err != nil {
		respond.Failure(c, h.log, err, "reading the overview failed")
		return
	}

	activity, err := h.describe(c, events)
	if err != nil {
		respond.Failure(c, h.log, err, "naming what the activity was about failed")
		return
	}

	c.JSON(http.StatusOK, overviewResponse{
		Days:      days,
		Counts:    counts,
		SignIns:   signIns,
		Daily:     daily,
		TopActors: actors,
		Activity:  activity,
	})
}

// Logs lists one page of the activity log, newest first, narrowed by the
// filters parseFilter reads, with the cursor the next page starts from.
func (h *Handler) Logs(c *gin.Context) {
	filter, err := parseFilter(c, query.Int(c, "limit", defaultLimit, maxLimit))
	if err != nil {
		respond.Failure(c, h.log, err, "reading the log filters failed")
		return
	}

	events, next, err := h.store.AuditQuery(c.Request.Context(), filter)
	if err != nil {
		respond.Failure(c, h.log, err, "listing logs failed")
		return
	}

	described, err := h.describe(c, events)
	if err != nil {
		respond.Failure(c, h.log, err, "naming what the logs were about failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": newLogResponses(events, described),
		"next": formatCursor(next),
	})
}

// Export writes the entries the filters match as a CSV file, newest first,
// up to exportLimit of them. Each row says what the logs page would: the
// same names hidden, the same detail left out, for the same administrator.
func (h *Handler) Export(c *gin.Context) {
	filter, err := parseFilter(c, exportPage)
	if err != nil {
		respond.Failure(c, h.log, err, "reading the log filters failed")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition",
		`attachment; filename="activity-`+time.Now().Format("2006-01-02")+`.csv"`)

	out := csv.NewWriter(c.Writer)
	_ = out.Write([]string{"time", "actor", "action", "target_type", "target_id", "target", "detail", "ip", "user_agent"})

	for written := 0; written < exportLimit; {
		events, next, err := h.store.AuditQuery(c.Request.Context(), filter)
		if err != nil {
			// The header is sent; the file ends here, and the log says why.
			h.log.Error("exporting logs failed", "error", err)
			break
		}

		described, err := h.describe(c, events)
		if err != nil {
			h.log.Error("naming what the export was about failed", "error", err)
			break
		}

		for i, event := range events {
			if written == exportLimit {
				break
			}

			row := described[i]
			targetType, targetID, target := "", "", ""
			if row.Target != nil {
				targetType, targetID, target = row.Target.Type, row.Target.ID, row.Target.Name
			}

			_ = out.Write([]string{
				event.CreatedAt.UTC().Format(time.RFC3339), row.Actor, row.Action,
				targetType, targetID, target, row.Detail, row.IP, event.UserAgent,
			})
			written++
		}

		out.Flush()
		if next == nil {
			break
		}
		filter.Before = next
	}

	out.Flush()
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
	case "login_flow":
		return admin.HasPermission(model.PermLoginFlowsRead)
	case "language":
		return admin.HasPermission(model.PermLanguagesRead)
	case "sso_connection":
		return admin.HasPermission(model.PermSSORead)
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
	case "sso_connection.sign_in_failed":
		// What the provider said, which is how a connection being set up
		// gets fixed — for whoever may read the connections.
		if admin := session.Admin(c); admin != nil && admin.HasPermission(model.PermSSORead) {
			return text("step") + ": " + text("reason")
		}
	case "language.translated":
		// Which of the two apps the text was for.
		if admin := session.Admin(c); admin != nil && admin.HasPermission(model.PermLanguagesRead) {
			return text("app")
		}
	}

	return ""
}
