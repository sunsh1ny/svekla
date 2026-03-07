package parser

import "errors"

var ErrEmptyQuery = errors.New("query is empty")
var ErrUnknownCommand = errors.New("unknown command")
var ErrInvalidArgumentsNumber = errors.New("invalid argument number")
