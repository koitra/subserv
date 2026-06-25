package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stephenafamo/bob"

	_ "github.com/jackc/pgx/v5/stdlib"

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

	mux := chi.NewMux()
	humaCfg := huma.DefaultConfig("Subserv", "0.1.0")
	hapi := humachi.New(mux, humaCfg)

	subsRepo := subscriptions.NewRepository(db)
	subsSvc := subscriptions.NewService(subsRepo, validate)
	subsHandler := subscriptions.NewHandler(subsSvc)
	subsHandler.Register(hapi)

	addr := net.JoinHostPort(cfg.HTTP.Host, strconv.FormatUint(uint64(cfg.HTTP.Port), 10))
	return http.ListenAndServe(addr, mux)
}
