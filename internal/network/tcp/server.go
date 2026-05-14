package tcp

import (
	"net"

	"go.uber.org/zap"
)

type Server struct {
	address        string
	maxConnections int
	handler        *Handler
	logger         *zap.Logger
	sem            chan struct{}
}

func NewServer(address string, maxConnections int, h *Handler, logger *zap.Logger) *Server {
	if maxConnections <= 0 {
		maxConnections = 1
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Server{
		address:        address,
		maxConnections: maxConnections,
		handler:        h,
		logger:         logger,
		sem:            make(chan struct{}, maxConnections),
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	return s.Serve(listener)
}

func (s *Server) Serve(listener net.Listener) error {
	defer listener.Close()

	s.logger.Info("tcp server started", zap.String("address", s.address))

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Warn("Error accepting connection", zap.Error(err))
			continue
		}

		if !s.acquireConnection() {
			s.rejectConnection(conn)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) acquireConnection() bool {
	select {
	case s.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.releaseConnection()
	defer s.recoverPanic()

	s.handler.Handle(conn)
}

func (s *Server) releaseConnection() {
	<-s.sem
}

func (s *Server) recoverPanic() {
	if r := recover(); r != nil {
		s.logger.Error("panic while handling connection", zap.Any("panic", r))
	}
}

func (s *Server) rejectConnection(conn net.Conn) {
	s.logger.Warn("max connections reached", zap.Int("limit", s.maxConnections))
	_ = conn.Close()
}
