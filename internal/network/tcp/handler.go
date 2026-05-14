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

const (
	maxInitialBufferSize = 1024

	responseErrorPrefix       = "ERR: "
	messageTooLargeResponse   = "message too large"
	connectionTimeoutResponse = "connection timeout"
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
	if logger == nil {
		logger = zap.NewNop()
	}
	if maxMessageSize <= 0 {
		maxMessageSize = maxInitialBufferSize
	}

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

	scanner := h.newScanner(conn)
	writer := newResponseWriter(conn, h.idleTimeout)

	for {
		if err := h.setReadDeadline(conn); err != nil {
			h.logger.Error("error setting read deadline", zap.Error(err))
			return
		}

		if !scanner.Scan() {
			h.handleReadError(scanner.Err(), writer)
			return
		}

		if err := h.handleMessage(scanner.Text(), writer); err != nil {
			h.logger.Error("error writing response", zap.Error(err))
			return
		}
	}
}

func (h *Handler) newScanner(conn net.Conn) *bufio.Scanner {
	initialBufferSize := h.initialBufferSize()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, initialBufferSize), h.maxMessageSize)

	return scanner
}

func (h *Handler) initialBufferSize() int {
	if h.maxMessageSize <= 0 {
		return maxInitialBufferSize
	}

	if h.maxMessageSize < maxInitialBufferSize {
		return h.maxMessageSize
	}

	return maxInitialBufferSize
}

func (h *Handler) handleMessage(message string, writer *responseWriter) error {
	query, err := h.parser.Parse(message)
	if err != nil {
		h.logger.Error("error parsing message", zap.Error(err))
		return writer.Error(err.Error())
	}

	result, err := h.service.Execute(query)
	if err != nil {
		h.logger.Error("error executing command", zap.Error(err))
		return writer.Error(err.Error())
	}

	return writer.Line(result)
}

func (h *Handler) handleReadError(err error, writer *responseWriter) {
	if err == nil {
		return
	}

	if errors.Is(err, bufio.ErrTooLong) {
		h.logger.Warn("message too large", zap.Int("max_message_size", h.maxMessageSize))
		h.writeTerminalError(writer, messageTooLargeResponse)
		return
	}

	if isTimeout(err) {
		h.logger.Warn("connection idle timeout reached", zap.Duration("idle_timeout", h.idleTimeout))
		h.writeTerminalError(writer, connectionTimeoutResponse)
		return
	}

	h.logger.Error("error reading message", zap.Error(err))
}

func (h *Handler) writeTerminalError(writer *responseWriter, message string) {
	if err := writer.Error(message); err != nil {
		h.logger.Error("error writing error response", zap.Error(err))
	}
}

func (h *Handler) setReadDeadline(conn net.Conn) error {
	if h.idleTimeout > 0 {
		return conn.SetReadDeadline(time.Now().Add(h.idleTimeout))
	}

	return nil
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

type responseWriter struct {
	conn    net.Conn
	writer  *bufio.Writer
	timeout time.Duration
}

func newResponseWriter(conn net.Conn, timeout time.Duration) *responseWriter {
	return &responseWriter{
		conn:    conn,
		writer:  bufio.NewWriter(conn),
		timeout: timeout,
	}
}

func (w *responseWriter) Error(message string) error {
	return w.Line(responseErrorPrefix + message)
}

func (w *responseWriter) Line(message string) error {
	if err := w.setDeadline(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w.writer, message); err != nil {
		return err
	}

	return w.writer.Flush()
}

func (w *responseWriter) setDeadline() error {
	if w.timeout <= 0 {
		return nil
	}

	return w.conn.SetWriteDeadline(time.Now().Add(w.timeout))
}
