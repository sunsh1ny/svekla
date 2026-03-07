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
	logger = logger.Named("engine")
	logger.Info("engine created")

	return &Engine{
		data:   make(map[string]string),
		logger: logger,
	}
}

func (e *Engine) Set(key string, value string) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyKey
	}

	e.mutex.Lock()
	e.data[key] = value
	e.mutex.Unlock()

	e.logger.Debug("set key", zap.String("key", key))

	return nil
}

func (e *Engine) Get(key string) (string, bool, error) {
	if strings.TrimSpace(key) == "" {
		return "", false, ErrEmptyKey
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
	if strings.TrimSpace(key) == "" {
		return false, ErrEmptyKey
	}

	e.mutex.Lock()
	_, ok := e.data[key]
	if !ok {
		e.mutex.Unlock()
		return false, nil
	}
	delete(e.data, key)
	e.mutex.Unlock()

	e.logger.Debug("delete key", zap.String("key", key))

	return true, nil
}
