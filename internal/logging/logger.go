package logging

import (
	"os"
	"path/filepath"
	"strings"

	"svekla/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg config.LoggingConfig) (*zap.Logger, error) {
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Encoding = "console"
	zapCfg.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	output := strings.TrimSpace(cfg.Output)
	switch output {
	case "", "stdout":
		zapCfg.OutputPaths = []string{"stdout"}
	case "stderr":
		zapCfg.OutputPaths = []string{"stderr"}
	default:
		dir := filepath.Dir(output)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, err
			}
		}
		zapCfg.OutputPaths = []string{output}
	}

	return zapCfg.Build()
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case config.LogLevelDebug:
		return zapcore.DebugLevel
	case config.LogLevelError:
		return zapcore.ErrorLevel
	case config.LogLevelInfo:
		fallthrough
	default:
		return zapcore.InfoLevel
	}
}
