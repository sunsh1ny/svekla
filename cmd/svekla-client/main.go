package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"svekla/internal/config"
)

const defaultConfigPath = "configs/local.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to yaml config")
	addressOverride := flag.String("addr", "", "tcp server address override")
	flag.Parse()

	if err := run(*configPath, *addressOverride, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "ERR:", err)
		os.Exit(1)
	}
}

func run(configPath string, addressOverride string, input io.Reader, output io.Writer) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	address := serverAddress(cfg, addressOverride)

	maxMessageSize, err := cfg.Network.ParsedMaxMessageSizeBytes()
	if err != nil {
		return fmt.Errorf("parse max message size: %w", err)
	}

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("dial server: %w", err)
	}
	defer conn.Close()

	return runInteractive(conn, input, output, cfg.Network.IdleTimeout, maxMessageSize)
}

func serverAddress(cfg config.Config, override string) string {
	if override != "" {
		return override
	}

	return cfg.Network.Address
}

func runInteractive(
	conn net.Conn,
	input io.Reader,
	output io.Writer,
	timeout time.Duration,
	maxMessageSize int,
) error {
	maxMessageSize = scannerMessageSizeLimit(maxMessageSize)

	stdin := bufio.NewScanner(input)
	stdin.Buffer(make([]byte, 0, scannerInitialBufferSize(maxMessageSize)), maxMessageSize)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		if _, err := fmt.Fprint(output, "> "); err != nil {
			return fmt.Errorf("write prompt: %w", err)
		}

		if !stdin.Scan() {
			break
		}

		line := strings.TrimSpace(stdin.Text())
		if line == "" {
			continue
		}

		response, err := sendRequest(conn, writer, reader, line, timeout)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(output, response); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}

	if err := stdin.Err(); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	return nil
}

func sendRequest(
	conn net.Conn,
	writer *bufio.Writer,
	reader *bufio.Reader,
	line string,
	timeout time.Duration,
) (string, error) {
	if err := setDeadline(timeout, conn.SetWriteDeadline); err != nil {
		return "", fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := fmt.Fprintln(writer, line); err != nil {
		return "", fmt.Errorf("write request: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("flush request: %w", err)
	}

	if err := setDeadline(timeout, conn.SetReadDeadline); err != nil {
		return "", fmt.Errorf("set read deadline: %w", err)
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	return strings.TrimRight(response, "\r\n"), nil
}

func setDeadline(timeout time.Duration, set func(time.Time) error) error {
	if timeout <= 0 {
		return nil
	}

	return set(time.Now().Add(timeout))
}

func scannerInitialBufferSize(maxMessageSize int) int {
	if maxMessageSize < 1024 {
		return maxMessageSize
	}

	return 1024
}

func scannerMessageSizeLimit(maxMessageSize int) int {
	if maxMessageSize <= 0 {
		return 1024
	}

	return maxMessageSize
}
