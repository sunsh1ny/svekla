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
	"svekla/internal/storage/wal"

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

	server, err := buildServer(cfg, maxMessageSize, logger)
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	if err := server.Run(); err != nil {
		logger.Error("server stopped with error", zap.Error(err))
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}

func buildServer(cfg config.Config, maxMessageSize int, logger *zap.Logger) (*tcp.Server, error) {
	p := parser.NewParser(logger)
	st := engine.NewEngine(logger)
	store, err := buildStore(cfg, st, logger)
	if err != nil {
		return nil, err
	}

	s := service.NewCommandService(store)
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

	return server, nil
}

func buildStore(cfg config.Config, st service.Store, logger *zap.Logger) (service.Store, error) {
	if !cfg.WAL.Enabled {
		return st, nil
	}

	segmentSize, err := cfg.WAL.ParsedSegmentSizeBytes()
	if err != nil {
		return nil, fmt.Errorf("parse wal segment_size: %w", err)
	}

	if err := wal.Recover(st, cfg.WAL.DataDirectory); err != nil {
		return nil, fmt.Errorf("recover wal: %w", err)
	}

	log, err := wal.Open(wal.Options{
		BatchSize:     cfg.WAL.FlushingBatchSize,
		BatchTimeout:  cfg.WAL.FlushingBatchTimeout,
		SegmentSize:   segmentSize,
		DataDirectory: cfg.WAL.DataDirectory,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("open wal: %w", err)
	}

	return wal.NewDurableStore(st, log)
}
