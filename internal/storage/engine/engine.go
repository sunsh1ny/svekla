package engine

import (
	"strings"
	"sync"

	"go.uber.org/zap"
)

type Engine struct {
	data   map[string]string
	mutex  sync.RWMutex
	logger *zap.Logger
}

func NewEngine(logger *zap.Logger) *Engine {
	if logger == nil {
		logger = zap.NewNop()
	}

	logger = logger.Named("engine")
	logger.Info("engine created")

	return &Engine{
		data:   make(map[string]string),
		logger: logger,
	}
}

func (e *Engine) Set(key string, value string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	e.set(key, value)

	e.logger.Debug("set key", zap.String("key", key))

	return nil
}

func (e *Engine) Get(key string) (string, bool, error) {
	if err := validateKey(key); err != nil {
		return "", false, err
	}

	e.mutex.RLock()
	value, ok := e.data[key]
	e.mutex.RUnlock()

	if !ok {
		return "", false, nil
	}

	e.logger.Debug("get key", zap.String("key", key))

	return value, true, nil
}

func (e *Engine) Delete(key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}

	deleted := e.delete(key)
	if deleted {
		e.logger.Debug("delete key", zap.String("key", key))
	}

	return deleted, nil
}

func (e *Engine) set(key string, value string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.data[key] = value
}

func (e *Engine) delete(key string) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	_, ok := e.data[key]
	if !ok {
		return false
	}
	delete(e.data, key)

	return true
}

func validateKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyKey
	}

	return nil
}
