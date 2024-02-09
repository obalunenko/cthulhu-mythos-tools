package config

import (
	"context"
	"errors"

	"github.com/obalunenko/getenv"
	"github.com/obalunenko/getenv/option"
	log "github.com/obalunenko/logger"
)

const (
	portEnv   = "PORT"
	hostEnv   = "HOST"
	levelEnv  = "LOG_LEVEL"
	formatEnv = "LOG_FORMAT"
)

type httpConfig struct {
	Port string `yaml:"port" json:"port"`
	Host string `yaml:"host" json:"host"`
}

type logConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
}

type Config struct {
	HTTP httpConfig `yaml:"http" json:"http"`
	Log  logConfig  `yaml:"log" json:"log"`
}

func DefaultConfig() *Config {
	return &Config{
		HTTP: httpConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Log: logConfig{
			Level:  "INFO",
			Format: "text",
		},
	}
}

func Load(ctx context.Context) (*Config, error) {
	return loadFromEnv(ctx)
}

func loadEnv[T string | []uint](ctx context.Context, key string, defaultVal T, opts ...option.Option) (T, error) {
	val, err := getenv.Env[T](key, opts...)
	if err != nil {
		if !errors.Is(err, getenv.ErrNotSet) {
			return val, err
		}

		log.WithFields(ctx, log.Fields{
			"env":     key,
			"default": defaultVal,
		}).Warn("Env not set - using default")

		val = defaultVal
	}

	return val, nil
}

func loadFromEnv(ctx context.Context) (*Config, error) {
	var errs error

	dflt := DefaultConfig()

	port, err := loadEnv[string](ctx, portEnv, dflt.HTTP.Port)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	host, err := loadEnv[string](ctx, hostEnv, dflt.HTTP.Host)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	level, err := loadEnv[string](ctx, levelEnv, dflt.Log.Level)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	format, err := loadEnv[string](ctx, formatEnv, dflt.Log.Format)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return &Config{
		HTTP: httpConfig{
			Port: port,
			Host: host,
		},
		Log: logConfig{
			Level:  level,
			Format: format,
		},
	}, nil
}
