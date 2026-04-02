package main

import (
	"fmt"
	"os"

	"svekla/internal/compute/parser"
	"svekla/internal/compute/service"
	"svekla/internal/config"
	"svekla/internal/logging"
	"svekla/internal/network/tcp"
	"svekla/internal/storage/engine"
)

func main() {
	cfg, err := config.Load("configs/local.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: load config:", err)
		os.Exit(1)
	}

	maxMessageSize, err := cfg.Network.ParsedMaxMessageSizeBytes()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: parse max_message_size:", err)
		os.Exit(1)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: init logger:", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync()
	}()

	p := parser.NewParser(logger.Named("parser"))
	st := engine.NewEngine(logger.Named("engine"))
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

	if err := server.Run(); err != nil {
		logger.Error("server stopped with error")
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
}
