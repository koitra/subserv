package main

import (
	"context"
	"errors"
	"fmt"
	"os"

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
	cfg.HTTP.EnableDocs = true

	app, err := app.New(cfg, validate)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	spec, err := app.Spec()
	if err != nil {
		return err
	}

	err = os.WriteFile("spec.yaml", spec, 0o644)
	if err != nil {
		return fmt.Errorf("write spec to spec.yaml: %w", err)
	}

	return nil
}
