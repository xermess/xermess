// Command xermess runs the authentication server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xermess/internal/api"
	"xermess/internal/config"
	"xermess/internal/database"
	"xermess/internal/mail"
	"xermess/internal/oidc"
	"xermess/internal/store"
	"xermess/locales"
)

// version and commit are stamped in at build time by `make build` and
// the Dockerfile, so a running server can say which build it is. A plain
// `go run` leaves them as they are.
var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(log); err != nil {
		log.Error("fatal", "error", err)
		os.Exit(1)
	}
}

// run is the whole startup, in order: read the configuration, open the
// database, apply migrations, then serve on top of a store. It is separate
// from main so every step can return an error instead of exiting from the
// middle of the startup.
func run(log *slog.Logger) error {
	log.Info("xermess starting", "version", version, "commit", commit)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Open(cfg.DB)
	if err != nil {
		return err
	}
	defer database.Close(db)

	log.Info("database connected")

	if cfg.DB.Migrate {
		if err := database.Migrate(db, cfg.DB, log); err != nil {
			return err
		}
	}

	// The store is the only thing that queries the database; the servers are
	// handed that rather than the connection itself.
	st := store.New(db)

	// What a fresh installation starts with for administrators' sign-ins,
	// from the configuration. An installation that already has the setting
	// keeps it: after the first start it is the panel's, not the file's.
	if err := st.EnsureAdminSecurity(context.Background(), cfg.AdminMFARequired); err != nil {
		return err
	}

	// The languages the server ships with: all of them on the first start,
	// and afterwards only the keys a release added. What an administrator has
	// written is never overwritten.
	shipped, err := locales.Shipped()
	if err != nil {
		return err
	}
	if err := st.EnsureLanguages(context.Background(), shipped); err != nil {
		return err
	}

	// The provider loads its signing keys, and makes any that are missing,
	// before anything is served: a wrong XERMESS_SECRET_KEY stops the server
	// here rather than failing the first sign-in.
	provider, err := oidc.New(context.Background(), cfg, st, mail.New(cfg.Mail, log), log)
	if err != nil {
		return err
	}

	public, err := api.NewPublic(cfg, log, provider)
	if err != nil {
		return err
	}

	admin, err := api.NewAdmin(cfg, st, log, provider)
	if err != nil {
		return err
	}

	// Signing keys rotate while the server runs; stopping the server stops it.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go provider.MaintainKeys(ctx)

	return serve(ctx, log, []*http.Server{
		{Addr: cfg.Addr, Handler: public, ReadHeaderTimeout: 10 * time.Second},
		{Addr: cfg.AdminAddr, Handler: admin, ReadHeaderTimeout: 10 * time.Second},
	})
}

// shutdownGrace is how long requests under way get to finish when the process
// is told to stop.
const shutdownGrace = 15 * time.Second

// serve runs the servers until one fails or the process is told to stop, then
// stops them all, letting requests under way finish. A container runtime sends
// SIGTERM before it kills a container, so a deploy does not cut sign-ins off
// half way.
func serve(ctx context.Context, log *slog.Logger, servers []*http.Server) error {
	failed := make(chan error, len(servers))
	for i, server := range servers {
		name := []string{"public", "admin"}[i]
		log.Info("server listening", "server", name, "addr", server.Addr)

		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				failed <- err
			}
		}()
	}

	var err error
	select {
	case err = <-failed:
	case <-ctx.Done():
		log.Info("shutting down")
	}

	shutdown, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()

	for _, server := range servers {
		_ = server.Shutdown(shutdown)
	}

	return err
}
