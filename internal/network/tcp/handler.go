package tcp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"svekla/internal/compute/parser"
	"svekla/internal/compute/service"
	"time"

	"go.uber.org/zap"
)

type Handler struct {
	maxMessageSize int
	idleTimeout    time.Duration
	parser         *parser.Parser
	service        *service.CommandService
	logger         *zap.Logger
}

func NewHandler(
	idleTimeout time.Duration,
	maxMessageSize int,
	p *parser.Parser,
	s *service.CommandService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		idleTimeout:    idleTimeout,
		maxMessageSize: maxMessageSize,
		parser:         p,
		service:        s,
		logger:         logger,
	}
}

func (h *Handler) Handle(conn net.Conn) {
	defer conn.Close()

	initialBufSize := h.maxMessageSize
	if initialBufSize > 1024 {
		initialBufSize = 1024
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, initialBufSize), h.maxMessageSize)

	writer := bufio.NewWriter(conn)

	for {
		if h.idleTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(h.idleTimeout)); err != nil {
				h.logger.Error("error setting read deadline", zap.Error(err))
				return
			}
		}

		if !scanner.Scan() {
			err := scanner.Err()
			if err == nil {
				return
			}

			if errors.Is(err, bufio.ErrTooLong) {
				h.logger.Warn("message too large", zap.Int("max_message_size", h.maxMessageSize))
				if writeErr := h.writeError(conn, writer, "message too large"); writeErr != nil {
					h.logger.Error("error writing oversize response", zap.Error(writeErr))
				}
				return
			}

			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				h.logger.Warn("connection idle timeout reached", zap.Duration("idle_timeout", h.idleTimeout))
				if writeErr := h.writeError(conn, writer, "connection timeout"); writeErr != nil {
					h.logger.Error("error writing timeout response", zap.Error(writeErr))
				}
				return
			}

			h.logger.Error("error reading message", zap.Error(err))
			return
		}

		line := scanner.Text()

		query, err := h.parser.Parse(line)
		if err != nil {
			h.logger.Error("error parsing message", zap.Error(err))
			if writeErr := h.writeError(conn, writer, err.Error()); writeErr != nil {
				h.logger.Error("error writing parse error response", zap.Error(writeErr))
				return
			}
			continue
		}

		result, err := h.service.Execute(query)
		if err != nil {
			h.logger.Error("error executing command", zap.Error(err))
			if writeErr := h.writeError(conn, writer, err.Error()); writeErr != nil {
				h.logger.Error("error writing execute error response", zap.Error(writeErr))
				return
			}
			continue
		}

		if err := h.writeLine(conn, writer, result); err != nil {
			h.logger.Error("error writing result", zap.Error(err))
			return
		}
	}
}

func (h *Handler) writeError(conn net.Conn, writer *bufio.Writer, message string) error {
	return h.writeLine(conn, writer, fmt.Sprintf("ERR: %s", message))
}

func (h *Handler) writeLine(conn net.Conn, writer *bufio.Writer, message string) error {
	if h.idleTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(h.idleTimeout)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(writer, message); err != nil {
		return err
	}

	return writer.Flush()
}
