package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	return path
}

func TestLoad_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")

	got, err := Load(path)

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}

	if !reflect.DeepEqual(got, Config{}) {
		t.Fatalf("got %#v, want %#v", got, Config{})
	}
}

func TestLoad_UnmarshalError(t *testing.T) {
	path := writeTempConfig(t, `
network:
  max_connections: abc
`)

	got, err := Load(path)

	var typeErr *yaml.TypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("expected yaml.TypeError, got %v", err)
	}

	if !reflect.DeepEqual(got, Config{}) {
		t.Fatalf("got %#v, want %#v", got, Config{})
	}
}

func TestLoad_Success(t *testing.T) {
	path := writeTempConfig(t, `
engine:
  type: in_memory

network:
  address: "127.0.0.1:9090"
  max_connections: 200
  max_message_size: "8KB"
  idle_timeout: 10s

logging:
  level: debug
  output: "/tmp/svekla.log"
`)

	got, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Config{
		Engine: EngineConfig{
			Type: EngineTypeInMemory,
		},
		Network: NetworkConfig{
			Address:        "127.0.0.1:9090",
			MaxConnections: 200,
			MaxMessageSize: "8KB",
			IdleTimeout:    10 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  LogLevelDebug,
			Output: "/tmp/svekla.log",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	path := writeTempConfig(t, `
network:
  address: "127.0.0.1:9000"

logging:
  level: error
`)

	got, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := DefaultConfig()
	want.Network.Address = "127.0.0.1:9000"
	want.Logging.Level = LogLevelError

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name:    "invalid engine type",
			cfg:     withConfig(func(cfg *Config) { cfg.Engine.Type = "postgres" }),
			wantErr: ErrInvalidEngineType,
		},
		{
			name:    "empty address",
			cfg:     withConfig(func(cfg *Config) { cfg.Network.Address = "" }),
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "invalid address format",
			cfg:     withConfig(func(cfg *Config) { cfg.Network.Address = "127.0.0.1" }),
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "invalid port",
			cfg:     withConfig(func(cfg *Config) { cfg.Network.Address = "127.0.0.1:70000" }),
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "invalid max connections",
			cfg:     withConfig(func(cfg *Config) { cfg.Network.MaxConnections = 0 }),
			wantErr: ErrInvalidMaxConnections,
		},
		{
			name:    "invalid log level",
			cfg:     withConfig(func(cfg *Config) { cfg.Logging.Level = "trace" }),
			wantErr: ErrInvalidLogLevel,
		},
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func withConfig(mutator func(cfg *Config)) Config {
	cfg := DefaultConfig()
	mutator(&cfg)

	return cfg
}
