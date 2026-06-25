package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/pressly/goose/v3"
	"github.com/stephenafamo/bob"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/koitra/subserv"
	"github.com/koitra/subserv/internal/config"
	"github.com/koitra/subserv/internal/subscriptions"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
		return
	}
}

func run() error {
	validate := validator.New()
	cfg, err := config.Load("config.yaml", validate)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	stdDB, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer func() { _ = stdDB.Close() }()
	stdDB.SetMaxOpenConns(cfg.DB.MaxConnections)
	db := bob.NewDB(stdDB)

	goose.SetBaseFS(subserv.MigrationsFS)
	err = goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("initialize sql migrations: %w", err)
	}

	err = goose.Up(stdDB, "migrations")
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	mux := chi.NewMux()
	humaCfg := huma.DefaultConfig("Subserv", "0.1.0")
	hapi := humachi.New(mux, humaCfg)

	subsRepo := subscriptions.NewRepository(db)
	subsSvc := subscriptions.NewService(subsRepo, validate)
	subsHandler := subscriptions.NewHandler(subsSvc)
	subsHandler.Register(hapi)

	addr := net.JoinHostPort(cfg.HTTP.Host, strconv.FormatUint(uint64(cfg.HTTP.Port), 10))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-sig

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
