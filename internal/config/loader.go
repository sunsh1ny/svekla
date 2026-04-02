package config

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("error parsing config file: %w", err)
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.Engine.Type != EngineTypeInMemory {
		return ErrInvalidEngineType
	}

	if cfg.Network.Address == "" {
		return ErrInvalidAddress
	}

	_, p, err := net.SplitHostPort(cfg.Network.Address)
	if err != nil {
		return ErrInvalidAddress
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return ErrInvalidAddress
	}

	if port < 1 || port > 65535 {
		return ErrInvalidAddress
	}

	if cfg.Network.MaxConnections <= 0 {
		return ErrInvalidMaxConnections
	}

	_, err = cfg.Network.ParsedMaxMessageSizeBytes()
	if err != nil {
		return ErrInvalidMaxMessageSize
	}

	if cfg.Network.IdleTimeout <= 0 {
		return ErrInvalidIdleTimeout
	}

	if cfg.Logging.Level != LogLevelInfo && cfg.Logging.Level != LogLevelError && cfg.Logging.Level != LogLevelDebug {
		return ErrInvalidLogLevel
	}

	return nil
}
