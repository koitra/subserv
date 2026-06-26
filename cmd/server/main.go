package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/koitra/subserv/internal/app"
	"github.com/koitra/subserv/internal/config"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
		return
	}
}

func run() error {
	cfgPath := os.Getenv("SUBSERV_CONFIG_PATH")
	validate := validator.New()
	cfg, err := config.Load(cfgPath, validate)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: cfg.App.Log.SlogLevel()},
			),
		),
	)

	app, err := app.New(cfg, validate)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()

	go func() {
		<-sig

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		slog.Info("Shutting down server")

		_ = app.Shutdown(ctx)
	}()

	slog.Info(
		"Starting server",
		slog.String("host", cfg.HTTP.Host),
		slog.Uint64("port", uint64(cfg.HTTP.Port)),
	)

	return app.ListenAndServe()
}
