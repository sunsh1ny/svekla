package wal

import "errors"

var ErrClosed = errors.New("wal is closed")
var ErrInvalidBatchSize = errors.New("invalid wal batch size")
var ErrInvalidBatchTimeout = errors.New("invalid wal batch timeout")
var ErrInvalidSegmentSize = errors.New("invalid wal segment size")
var ErrInvalidDataDirectory = errors.New("invalid wal data directory")
var ErrInvalidRecord = errors.New("invalid wal record")
