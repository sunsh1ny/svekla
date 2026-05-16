package wal

import (
	"strings"
	"time"
)

type Options struct {
	BatchSize     int
	BatchTimeout  time.Duration
	SegmentSize   int
	DataDirectory string
}

func (o Options) validate() error {
	if o.BatchSize <= 0 {
		return ErrInvalidBatchSize
	}

	if o.BatchTimeout <= 0 {
		return ErrInvalidBatchTimeout
	}

	if o.SegmentSize <= 0 {
		return ErrInvalidSegmentSize
	}

	if strings.TrimSpace(o.DataDirectory) == "" {
		return ErrInvalidDataDirectory
	}

	return nil
}
