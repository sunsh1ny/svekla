package main

import (
	"bufio"
	"flag"
	"fmt"
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

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: load config:", err)
		os.Exit(1)
	}

	address := cfg.Network.Address
	if *addressOverride != "" {
		address = *addressOverride
	}

	maxMessageSize, err := cfg.Network.ParsedMaxMessageSizeBytes()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: parse max message size:", err)
		os.Exit(1)
	}

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: dial server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	initialBufSize := maxMessageSize
	if initialBufSize > 1024 {
		initialBufSize = 1024
	}

	stdin := bufio.NewScanner(os.Stdin)
	stdin.Buffer(make([]byte, 0, initialBufSize), maxMessageSize)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		fmt.Print("> ")

		if !stdin.Scan() {
			break
		}

		line := strings.TrimSpace(stdin.Text())
		if line == "" {
			continue
		}

		if cfg.Network.IdleTimeout > 0 {
			if err := conn.SetWriteDeadline(time.Now().Add(cfg.Network.IdleTimeout)); err != nil {
				fmt.Fprintln(os.Stderr, "ERR: set write deadline:", err)
				os.Exit(1)
			}
		}

		if _, err := fmt.Fprintln(writer, line); err != nil {
			fmt.Fprintln(os.Stderr, "ERR: write request:", err)
			os.Exit(1)
		}

		if err := writer.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, "ERR: flush request:", err)
			os.Exit(1)
		}

		if cfg.Network.IdleTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(cfg.Network.IdleTimeout)); err != nil {
				fmt.Fprintln(os.Stderr, "ERR: set read deadline:", err)
				os.Exit(1)
			}
		}

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERR: read response:", err)
			os.Exit(1)
		}

		fmt.Println(strings.TrimRight(response, "\r\n"))
	}

	if err := stdin.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "ERR: read stdin:", err)
		os.Exit(1)
	}
}
