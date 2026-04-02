package config

import "time"

func DefaultConfig() Config {
	return Config{
		Engine: EngineConfig{
			Type: EngineTypeInMemory,
		},
		Network: NetworkConfig{
			Address:        "127.0.0.1:3223",
			MaxConnections: 100,
			MaxMessageSize: "4KB",
			IdleTimeout:    5 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:  LogLevelInfo,
			Output: "stdout",
		},
	}
}
