package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

type (
	Full struct {
		DB   DB   `koanf:"db"`
		HTTP HTTP `koanf:"http"`
		App  App  `koanf:"app"`
	}

	App struct {
		Log LogLevel `koanf:"log"`
	}

	DB struct {
		URL            string `koanf:"url"            validate:"required,url"`
		MaxConnections int    `koanf:"maxconnections" validate:"omitempty,gte=1"`
	}

	HTTP struct {
		Host       string `koanf:"host" validate:"required"`
		Port       uint16 `koanf:"port" validate:"required"`
		EnableDocs bool   `koanf:"docs"`
	}
)

func Load(cfgPath string, validate *validator.Validate) (Full, error) {
	k := koanf.New(".")

	err := k.Load(structs.Provider(new(defaultConfig()), "koanf"), nil)
	if err != nil {
		return Full{}, fmt.Errorf("load default: %w", err)
	}

	if cfgPath != "" {
		f, err := os.Open(cfgPath)
		if err == nil {
			defer func() { _ = f.Close() }()

			err = k.Load(file.Provider(cfgPath), yaml.Parser())
			if err != nil {
				return Full{}, fmt.Errorf("load yaml: %w", err)
			}
			slog.Debug("Loaded config from file", slog.String("config", cfgPath))
		} else if os.IsNotExist(err) {
			slog.Warn("Provided config file was not found", slog.String("config", cfgPath))
		} else {
			return Full{}, fmt.Errorf("open config file: %w", err)
		}
	}

	err = k.Load(env.Provider(".", env.Opt{
		Prefix: "SUBSERV_",
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "SUBSERV_")), "_", ".")
			if strings.Contains(v, " ") {
				return k, strings.Split(v, " ")
			}

			return k, v
		},
	}), nil)
	if err != nil {
		return Full{}, fmt.Errorf("load env config: %w", err)
	}
	slog.Debug("Loaded config from env")

	cfg := Full{}
	err = k.Unmarshal("", &cfg)
	if err != nil {
		return Full{}, fmt.Errorf("unmarshal config: %w", err)
	}

	err = validate.Struct(&cfg)
	if err != nil {
		return Full{}, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func defaultConfig() Full {
	return Full{
		DB:   DB{MaxConnections: 50},
		App:  App{Log: Info},
		HTTP: HTTP{Host: "localhost", Port: 33220},
	}
}

type LogLevel string

const (
	Warn  LogLevel = "warn"
	Info  LogLevel = "info"
	Debug LogLevel = "debug"
)

func (l *LogLevel) UnmarshalText(text []byte) error {
	v := string(text)
	switch LogLevel(v) {
	case Warn, Info, Debug:
		*l = LogLevel(v)
		return nil
	default:
		return fmt.Errorf("unknown log level: %s", v)
	}
}

func (l LogLevel) SlogLevel() slog.Level {
	switch l {
	case Warn:
		return slog.LevelWarn
	case Info:
		return slog.LevelInfo
	case Debug:
		return slog.LevelDebug
	}

	return slog.LevelInfo
}
