package wal

import "fmt"

type Store interface {
	Set(key string, value string) error
	Get(key string) (string, bool, error)
	Delete(key string) (bool, error)
}

type DurableStore struct {
	store Store
	log   *Log
}

func NewDurableStore(store Store, log *Log) (*DurableStore, error) {
	if store == nil {
		return nil, fmt.Errorf("store is nil")
	}
	if log == nil {
		return nil, fmt.Errorf("wal log is nil")
	}

	return &DurableStore{
		store: store,
		log:   log,
	}, nil
}

func Recover(store Store, dataDirectory string) error {
	if store == nil {
		return fmt.Errorf("store is nil")
	}

	return Replay(dataDirectory, func(record Record) error {
		switch record.Operation {
		case OperationSet:
			return store.Set(record.Key, record.Value)
		case OperationDel:
			_, err := store.Delete(record.Key)
			return err
		default:
			return ErrInvalidRecord
		}
	})
}

func (s *DurableStore) Set(key string, value string) error {
	if _, _, err := s.store.Get(key); err != nil {
		return err
	}

	if err := s.log.Append(NewSetRecord(key, value)); err != nil {
		return err
	}

	return s.store.Set(key, value)
}

func (s *DurableStore) Get(key string) (string, bool, error) {
	return s.store.Get(key)
}

func (s *DurableStore) Delete(key string) (bool, error) {
	_, ok, err := s.store.Get(key)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	if err := s.log.Append(NewDeleteRecord(key)); err != nil {
		return false, err
	}

	return s.store.Delete(key)
}
