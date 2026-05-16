package wal

import (
	"encoding/json"
	"fmt"
)

const (
	OperationSet = "SET"
	OperationDel = "DEL"
)

type Record struct {
	Operation string `json:"op"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
}

func NewSetRecord(key string, value string) Record {
	return Record{
		Operation: OperationSet,
		Key:       key,
		Value:     value,
	}
}

func NewDeleteRecord(key string) Record {
	return Record{
		Operation: OperationDel,
		Key:       key,
	}
}

func (r Record) Validate() error {
	switch r.Operation {
	case OperationSet, OperationDel:
	default:
		return fmt.Errorf("%w: unknown operation %q", ErrInvalidRecord, r.Operation)
	}

	if r.Key == "" {
		return fmt.Errorf("%w: empty key", ErrInvalidRecord)
	}

	return nil
}

func encodeRecord(record Record) ([]byte, error) {
	if err := record.Validate(); err != nil {
		return nil, err
	}

	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal wal record: %w", err)
	}

	return append(data, '\n'), nil
}

func decodeRecord(data []byte) (Record, error) {
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, fmt.Errorf("%w: %v", ErrInvalidRecord, err)
	}

	if err := record.Validate(); err != nil {
		return Record{}, err
	}

	return record, nil
}
