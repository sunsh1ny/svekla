package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"svekla/internal/compute/parser"
	"svekla/internal/storage/engine"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	p := parser.NewParser(logger)
	st := engine.NewEngine(logger)

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		fmt.Fprint(writer, "> ")
		writer.Flush()

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		q, err := p.Parse(line)
		if err != nil {
			fmt.Fprintln(writer, "ERR:", err)
			continue
		}

		switch q.CommandID() {
		case parser.SetCommandID:
			args := q.Arguments()
			err = st.Set(args[0], args[1])
			if err != nil {
				fmt.Fprintln(writer, "ERR:", err)
				continue
			}
			fmt.Fprintln(writer, "OK")

		case parser.GetCommandID:
			args := q.Arguments()
			value, ok, err := st.Get(args[0])
			if err != nil {
				fmt.Fprintln(writer, "ERR:", err)
				continue
			}
			if !ok {
				fmt.Fprintln(writer, "NOT FOUND")
				continue
			}
			fmt.Fprintln(writer, value)

		case parser.DelCommandID:
			args := q.Arguments()
			deleted, err := st.Delete(args[0])
			if err != nil {
				fmt.Fprintln(writer, "ERR:", err)
				continue
			}
			if !deleted {
				fmt.Fprintln(writer, "NOT FOUND")
				continue
			}
			fmt.Fprintln(writer, "OK")

		default:
			fmt.Fprintln(writer, "ERR:", errors.New("unsupported command"))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
	}
}
