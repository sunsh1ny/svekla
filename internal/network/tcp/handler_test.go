package tcp

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"svekla/internal/compute/parser"
	"svekla/internal/compute/service"
	"svekla/internal/storage/engine"

	"go.uber.org/zap"
)

func newTestHandler(t *testing.T, idleTimeout time.Duration, maxMessageSize int) *Handler {
	t.Helper()

	logger := zap.NewNop()
	p := parser.NewParser(logger)
	st := engine.NewEngine(logger)
	s := service.NewCommandService(st)

	return NewHandler(idleTimeout, maxMessageSize, p, s, logger)
}

func startHandler(t *testing.T, idleTimeout time.Duration, maxMessageSize int) (net.Conn, *bufio.Reader, func()) {
	t.Helper()

	serverConn, clientConn := net.Pipe()
	handler := newTestHandler(t, idleTimeout, maxMessageSize)

	done := make(chan struct{})
	go func() {
		defer close(done)
		handler.Handle(serverConn)
	}()

	reader := bufio.NewReader(clientConn)

	cleanup := func() {
		_ = clientConn.Close()
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("handler did not stop in time")
		}
	}

	return clientConn, reader, cleanup
}

func sendCommand(t *testing.T, conn net.Conn, reader *bufio.Reader, command string) string {
	t.Helper()

	if !strings.HasSuffix(command, "\n") {
		command += "\n"
	}

	if err := conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fatalf("set write deadline: %v", err)
	}

	if _, err := conn.Write([]byte(command)); err != nil {
		t.Fatalf("write command: %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return strings.TrimRight(resp, "\r\n")
}

func freeTCPAddress(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on ephemeral port: %v", err)
	}
	defer l.Close()

	return l.Addr().String()
}

func waitForServer(t *testing.T, address string) {
	t.Helper()

	deadline := time.Now().Add(1 * time.Second)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("server on %s did not start in time", address)
}

func TestHandler_HappyPath(t *testing.T) {
	conn, reader, cleanup := startHandler(t, 1*time.Second, 1024)
	defer cleanup()

	if got := sendCommand(t, conn, reader, "SET key value"); got != "OK" {
		t.Fatalf("SET response = %q, want %q", got, "OK")
	}

	if got := sendCommand(t, conn, reader, "GET key"); got != "value" {
		t.Fatalf("GET response = %q, want %q", got, "value")
	}

	if got := sendCommand(t, conn, reader, "DEL key"); got != "OK" {
		t.Fatalf("DEL response = %q, want %q", got, "OK")
	}

	if got := sendCommand(t, conn, reader, "GET key"); got != "NOT_FOUND" {
		t.Fatalf("GET missing response = %q, want %q", got, "NOT_FOUND")
	}
}

func TestHandler_ParseError(t *testing.T) {
	conn, reader, cleanup := startHandler(t, 1*time.Second, 1024)
	defer cleanup()

	got := sendCommand(t, conn, reader, "LOL kek")
	want := "ERR: unknown command"

	if got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}

func TestHandler_NotFound(t *testing.T) {
	conn, reader, cleanup := startHandler(t, 1*time.Second, 1024)
	defer cleanup()

	got := sendCommand(t, conn, reader, "GET missing")
	want := "NOT_FOUND"

	if got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}

func TestHandler_MaxMessageSize(t *testing.T) {
	logger := zap.NewNop()
	p := parser.NewParser(logger)
	st := engine.NewEngine(logger)
	s := service.NewCommandService(st)
	h := NewHandler(1*time.Second, 8, p, s, logger)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			return
		}

		h.Handle(conn)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	if _, err := conn.Write([]byte("GET key12\n")); err != nil {
		t.Fatalf("write oversized command: %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	got := strings.TrimRight(resp, "\r\n")
	want := "ERR: message too large"

	if got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}

	<-done
}

func TestHandler_IdleTimeout(t *testing.T) {
	conn, reader, cleanup := startHandler(t, 50*time.Millisecond, 1024)
	defer cleanup()

	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read timeout response: %v", err)
	}

	got := strings.TrimRight(resp, "\r\n")
	want := "ERR: connection timeout"

	if got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}

func TestServer_MaxConnections(t *testing.T) {
	logger := zap.NewNop()
	p := parser.NewParser(logger)
	st := engine.NewEngine(logger)
	s := service.NewCommandService(st)
	h := NewHandler(2*time.Second, 1024, p, s, logger)

	address := freeTCPAddress(t)
	server := NewServer(address, 1, h, logger)

	go func() {
		_ = server.Run()
	}()

	waitForServer(t, address)

	conn1, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("dial first connection: %v", err)
	}
	defer conn1.Close()

	conn2, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("dial second connection: %v", err)
	}
	defer conn2.Close()

	if err := conn2.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		t.Fatalf("set read deadline on second connection: %v", err)
	}

	var buf [1]byte
	_, err = conn2.Read(buf[:])
	if err == nil {
		t.Fatal("expected second connection to be closed when max connections reached")
	}
}
