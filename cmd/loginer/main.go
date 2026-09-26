// Command loginer runs the authentication server.
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

	"loginer/i18n"
	"loginer/internal/api"
	"loginer/internal/brand"
	"loginer/internal/cache"
	"loginer/internal/config"
	"loginer/internal/database"
	"loginer/internal/jose"
	"loginer/internal/mail"
	"loginer/internal/oidc"
	"loginer/internal/store"
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
	log.Info(brand.Name+" starting", "version", version, "commit", commit)

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

	// Redis, when one is configured: the cache in front of the reads every
	// page makes, and where the rate limit counts. A configured Redis that
	// does not answer stops the server here, like a database that does not;
	// none configured runs without, as the server always could.
	shared, err := cache.Open(context.Background(), cfg.Redis, log)
	if err != nil {
		return err
	}
	defer shared.Close()

	if shared != nil {
		log.Info("redis connected", "addr", cfg.Redis.Addr(), "db", cfg.Redis.DB, "prefix", cfg.Redis.Prefix)
	} else {
		log.Info("redis is not configured; reading every page from the database")
	}

	// The store is the only thing that queries the database; the servers are
	// handed that rather than the connection itself.
	st := store.New(db).WithCache(shared)

	// The secret key, which seals what the database must not hold in the
	// clear: the mail server's password here, and the signing keys in the
	// provider below.
	sealer, err := jose.NewSealer(cfg.SecretKey)
	if err != nil {
		return err
	}

	// What a fresh installation starts with for administrators' sign-ins, for
	// sending email, and for the codes it emails, from the configuration. An
	// installation that already has any of these keeps it: after the first
	// start they are the panel's, not the file's.
	if err := st.EnsureAdminSecurity(context.Background(), cfg.AdminMFARequired); err != nil {
		return err
	}
	if err := ensureMailSettings(context.Background(), st, sealer, cfg.Mail); err != nil {
		return err
	}
	if err := st.EnsureOTPSettings(context.Background()); err != nil {
		return err
	}

	// The languages the server ships with: all of them on the first start,
	// and afterwards only the keys a release added. What an administrator has
	// written is never overwritten.
	shipped, err := i18n.Shipped()
	if err != nil {
		return err
	}
	if err := st.EnsureLanguages(context.Background(), shipped); err != nil {
		return err
	}

	// Email goes through whichever server the Mail page names when a message
	// is sent, rather than whichever one the configuration named at startup.
	mailer := mail.New(mail.FromStore(st, sealer), log)

	// The provider loads its signing keys, and makes any that are missing,
	// before anything is served: a wrong LOGINER_SECRET_KEY stops the server
	// here rather than failing the first sign-in.
	provider, err := oidc.New(context.Background(), cfg, st, mailer, log)
	if err != nil {
		return err
	}

	public, err := api.NewPublic(cfg, log, provider, shared)
	if err != nil {
		return err
	}

	admin, err := api.NewAdmin(cfg, st, log, provider, shared)
	if err != nil {
		return err
	}

	// Signing keys rotate, and expired records are swept, while the server
	// runs; stopping the server stops both.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go provider.MaintainKeys(ctx)
	go st.KeepSwept(ctx, cfg.AuditRetention, log)

	return serve(ctx, log, []*http.Server{
		listener(cfg.Addr, public),
		listener(cfg.AdminAddr, admin),
	})
}

// ensureMailSettings writes the mail server a fresh installation starts with,
// from LOGINER_SMTP_*, sealing the password with the secret key the way the
// panel does when it saves one.
func ensureMailSettings(ctx context.Context, st *store.Store, sealer *jose.Sealer, cfg config.Mail) error {
	settings, password := mail.Seed(cfg)

	if password != "" {
		sealed, err := sealer.SealBytes([]byte(password))
		if err != nil {
			return err
		}
		settings.Password = sealed
	}

	return st.EnsureMailSettings(ctx, settings)
}

// What a connection is given before it is cut off. Without these a caller can
// hold a connection, and the goroutine serving it, for as long as it likes by
// sending a byte now and then — or by reading an answer that slowly.
//
// The header and the body have short limits: no endpoint here takes an upload,
// and the longest body anybody sends is a language's text.
//
// Writing gets much longer because answering can mean calling somebody else.
// A sign-in through an identity provider reads its discovery document, trades
// the code at its token endpoint, fetches its keys, and may read its userinfo
// — four calls of up to socialTimeout each — and a sign-in that has to send a
// verification email waits for the mail server. Two minutes is well past all
// of that together, and still a bound where there was none.
const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 2 * time.Minute
	idleTimeout       = 2 * time.Minute
)

// listener is one of the two servers, with the timeouts both are held to.
func listener(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
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
