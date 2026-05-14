package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

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
	if err := validateEngine(cfg.Engine); err != nil {
		return err
	}

	if err := validateNetwork(cfg.Network); err != nil {
		return err
	}

	return validateLogging(cfg.Logging)
}

func validateEngine(cfg EngineConfig) error {
	if cfg.Type != EngineTypeInMemory {
		return ErrInvalidEngineType
	}

	return nil
}

func validateNetwork(cfg NetworkConfig) error {
	if err := validateAddress(cfg.Address); err != nil {
		return err
	}

	if cfg.MaxConnections <= 0 {
		return ErrInvalidMaxConnections
	}

	if _, err := cfg.ParsedMaxMessageSizeBytes(); err != nil {
		return ErrInvalidMaxMessageSize
	}

	if cfg.IdleTimeout <= 0 {
		return ErrInvalidIdleTimeout
	}

	return nil
}

func validateAddress(address string) error {
	if address == "" {
		return ErrInvalidAddress
	}

	_, p, err := net.SplitHostPort(address)
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

	return nil
}

func validateLogging(cfg LoggingConfig) error {
	if !isValidLogLevel(cfg.Level) {
		return ErrInvalidLogLevel
	}

	return nil
}

func isValidLogLevel(level string) bool {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case LogLevelInfo, LogLevelError, LogLevelDebug:
		return true
	default:
		return false
	}
}
