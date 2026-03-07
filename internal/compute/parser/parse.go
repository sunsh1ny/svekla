package parser

import (
	"strings"

	"go.uber.org/zap"
)

type Parser struct {
	logger *zap.Logger
}

func NewParser(logger *zap.Logger) *Parser {
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
	commandID := commandNameToCommandID(commandName)
	if commandID == UnknownCommandID {
		return Query{}, ErrUnknownCommand
	}

	arguments := fields[1:]
	commandArgumentsNum := commandArgumentsCount(commandID)

	if len(arguments) != commandArgumentsNum {
		return Query{}, ErrInvalidArgumentsNumber
	}

	return Query{
		commandID: commandID,
		arguments: arguments,
	}, nil
}
