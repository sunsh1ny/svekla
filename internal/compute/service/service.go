package service

import (
	"svekla/internal/compute/parser"
)

type Store interface {
	Set(key string, value string) error
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
}

type CommandService struct {
	store Store
}

func NewCommandService(store Store) *CommandService {
	return &CommandService{
		store: store,
	}
}

func (s *CommandService) Execute(query parser.Query) (string, error) {
	switch query.CommandID() {
	case parser.SetCommandID:
		return s.set(query)
	case parser.GetCommandID:
		return s.get(query)
	case parser.DelCommandID:
		return s.delete(query)
	default:
		return "", ErrUnknownCommand
	}
}

func (s *CommandService) set(query parser.Query) (string, error) {
	key, err := commandArgument(query, 0)
	if err != nil {
		return "", err
	}

	value, err := commandArgument(query, 1)
	if err != nil {
		return "", err
	}

	if err := s.store.Set(key, value); err != nil {
		return "", err
	}

	return ResultOK, nil
}

func (s *CommandService) get(query parser.Query) (string, error) {
	key, err := commandArgument(query, 0)
	if err != nil {
		return "", err
	}

	value, ok, err := s.store.Get(key)
	if err != nil {
		return "", err
	}

	if !ok {
		return ResultNotFound, nil
	}

	return value, nil
}

func (s *CommandService) delete(query parser.Query) (string, error) {
	key, err := commandArgument(query, 0)
	if err != nil {
		return "", err
	}

	ok, err := s.store.Delete(key)
	if err != nil {
		return "", err
	}

	if !ok {
		return ResultNotFound, nil
	}

	return ResultOK, nil
}

func commandArgument(query parser.Query, index int) (string, error) {
	argument, ok := query.Argument(index)
	if !ok {
		return "", ErrInvalidCommandArguments
	}

	return argument, nil
}
