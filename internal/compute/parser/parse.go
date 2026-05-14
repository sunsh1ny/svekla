package parser

import (
	"strings"

	"go.uber.org/zap"
)

type Parser struct {
	logger *zap.Logger
}

func NewParser(logger *zap.Logger) *Parser {
	if logger == nil {
		logger = zap.NewNop()
	}

	logger = logger.Named("parser")
	logger.Debug("initializing parser")

	return &Parser{
		logger: logger,
	}
}

func (parser *Parser) Parse(raw string) (Query, error) {
	fields := strings.Fields(raw)

	if len(fields) == 0 {
		return Query{}, ErrEmptyQuery
	}

	commandName := fields[0]
	command, ok := commandByName(commandName)
	if !ok {
		return Query{}, ErrUnknownCommand
	}

	arguments := fields[1:]
	if len(arguments) != command.argumentCount {
		return Query{}, ErrInvalidArgumentsNumber
	}

	return NewQuery(command.id, arguments), nil
}
