package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/pressly/goose/v3"
	"github.com/stephenafamo/bob"
	"go.yaml.in/yaml/v3"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/koitra/subserv"
	"github.com/koitra/subserv/internal/config"
	"github.com/koitra/subserv/internal/humaext"
	"github.com/koitra/subserv/internal/subscriptions"
)

type App struct {
	http *http.Server
	hapi huma.API
}

func New(cfg config.Full, validate *validator.Validate) (App, error) {
	stdDB, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		return App{}, fmt.Errorf("connect to db: %w", err)
	}
	stdDB.SetMaxOpenConns(cfg.DB.MaxConnections)
	db := bob.NewDB(stdDB)

	goose.SetBaseFS(subserv.MigrationsFS)
	err = goose.SetDialect("postgres")
	if err != nil {
		return App{}, fmt.Errorf("initialize sql migrations: %w", err)
	}

	err = goose.Up(stdDB, "migrations")
	if err != nil {
		return App{}, fmt.Errorf("migrate database: %w", err)
	}

	mux := chi.NewMux()
	humaCfg := huma.DefaultConfig("Subserv", "0.1.0")
	if !cfg.HTTP.EnableDocs {
		humaCfg.DocsPath = ""
		humaCfg.SchemasPath = ""
		humaCfg.OpenAPIPath = ""
	}

	hapi := humachi.New(mux, humaCfg)
	hapi.UseMiddleware(humaext.SlogMiddleware)

	subsRepo := subscriptions.NewRepository(db)
	subsSvc := subscriptions.NewService(subsRepo, validate)
	subsHandler := subscriptions.NewHandler(subsSvc)
	subsHandler.Register(hapi)

	addr := net.JoinHostPort(cfg.HTTP.Host, strconv.FormatUint(uint64(cfg.HTTP.Port), 10))

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	app := App{
		http: srv,
		hapi: hapi,
	}

	return app, nil
}

func (a *App) Spec() ([]byte, error) {
	spec := a.hapi.OpenAPI()
	data, err := yaml.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	return data, nil
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.http.Shutdown(ctx)
}

func (a *App) ListenAndServe() error {
	err := a.http.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
