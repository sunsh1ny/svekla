package service

import "errors"

var ErrUnknownCommand = errors.New("unknown command")
var ErrInvalidCommandArguments = errors.New("invalid command arguments")
