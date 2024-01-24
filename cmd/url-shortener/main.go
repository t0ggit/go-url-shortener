package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/update"
	middlewareLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/prettyslog"
	"url-shortener/internal/storage/sqlite"
)

func main() {
	// init config (cleanenv)
	cfg := config.MustLoad()

	// init logger (slog)
	log := setupLogger(cfg.Env)
	log.Info("starting url-shortener", slog.String("env", cfg.Env))

	// init storage (sqlite, but postgres in future)
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("cannot init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	_ = storage

	// init router (chi with render)
	router := chi.NewRouter()
	// middlewares for router
	router.Use(
		middleware.RequestID,
		middleware.RealIP,
		middlewareLogger.New(log),
		middleware.Recoverer,
		middleware.URLFormat,
	)

	router.Get("/{alias}", redirect.New(log, storage))

	router.Route("/modify", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/save-url", save.New(log, storage))
		r.Post("/delete-url", delete.New(log, storage))
		r.Post("/update-url", update.New(log, storage))
	})

	// run server
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
	}

	// OS Signal handling
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		sig := <-sigchan
		log.Info("shutting down server", slog.String("signal", sig.String()))
		_ = server.Close()
	}()

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("cannot start server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Error("server stopped", slog.String("error", err.Error()))
}

const (
	envLocal = "local"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		//log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		log = setupPrettySlog()
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	options := prettyslog.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := options.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
