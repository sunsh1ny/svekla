package service

import (
	"svekla/internal/compute/parser"
	"svekla/internal/storage/engine"
)

type CommandService struct {
	engine *engine.Engine
}

func NewCommandService(engine *engine.Engine) *CommandService {
	return &CommandService{
		engine: engine,
	}
}

func (service *CommandService) Execute(query parser.Query) (string, error) {
	args := query.Arguments()
	commandID := query.CommandID()

	switch commandID {
	case parser.SetCommandID:
		err := service.engine.Set(args[0], args[1])
		if err != nil {
			return "", err
		}

		return ResultOK, nil
	case parser.GetCommandID:
		val, ok, err := service.engine.Get(args[0])
		if err != nil {
			return "", err
		}

		if !ok {
			return ResultNotFound, nil
		}

		return val, nil
	case parser.DelCommandID:
		ok, err := service.engine.Delete(args[0])
		if err != nil {
			return "", err
		}

		if !ok {
			return ResultNotFound, nil
		}

		return ResultOK, nil
	default:
		return "", ErrUnknownCommand
	}
}
