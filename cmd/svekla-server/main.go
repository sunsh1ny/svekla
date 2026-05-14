package main

import (
	"flag"
	"fmt"
	"os"

	"svekla/internal/compute/parser"
	"svekla/internal/compute/service"
	"svekla/internal/config"
	"svekla/internal/logging"
	"svekla/internal/network/tcp"
	"svekla/internal/storage/engine"

	"go.uber.org/zap"
)

const defaultConfigPath = "configs/local.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to yaml config")
	flag.Parse()

	if err := run(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	maxMessageSize, err := cfg.Network.ParsedMaxMessageSizeBytes()
	if err != nil {
		return fmt.Errorf("parse max_message_size: %w", err)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	server := buildServer(cfg, maxMessageSize, logger)
	if err := server.Run(); err != nil {
		logger.Error("server stopped with error", zap.Error(err))
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}

func buildServer(cfg config.Config, maxMessageSize int, logger *zap.Logger) *tcp.Server {
	p := parser.NewParser(logger)
	st := engine.NewEngine(logger)
	s := service.NewCommandService(st)
	h := tcp.NewHandler(
		cfg.Network.IdleTimeout,
		maxMessageSize,
		p,
		s,
		logger.Named("tcp_handler"),
	)
	server := tcp.NewServer(
		cfg.Network.Address,
		cfg.Network.MaxConnections,
		h,
		logger.Named("tcp_server"),
	)

	return server
}
