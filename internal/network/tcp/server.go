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
	defer listener.Close()

	s.logger.Info("tcp server started", zap.String("address", s.address))

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Warn("Error accepting connection", zap.Error(err))
			continue
		}

		select {
		case s.sem <- struct{}{}:
			go func() {
				defer func() {
					<-s.sem

					if r := recover(); r != nil {
						s.logger.Error("panic while handling connection", zap.Any("panic", r))
					}
				}()

				s.handler.Handle(conn)
			}()
		default:
			s.logger.Warn("max connections reached", zap.Int("limit", s.maxConnections))
			_ = conn.Close()
		}
	}
}
