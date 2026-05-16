package config

import "errors"

var ErrInvalidEngineType = errors.New("invalid engine type")
var ErrInvalidAddress = errors.New("invalid network address")
var ErrInvalidMaxConnections = errors.New("invalid max connections")
var ErrInvalidMaxMessageSize = errors.New("invalid max message size")
var ErrInvalidIdleTimeout = errors.New("invalid idle timeout")
var ErrInvalidLogLevel = errors.New("invalid log level")
var ErrInvalidWALBatchSize = errors.New("invalid wal flushing batch size")
var ErrInvalidWALBatchTimeout = errors.New("invalid wal flushing batch timeout")
var ErrInvalidWALSegmentSize = errors.New("invalid wal segment size")
var ErrInvalidWALDataDirectory = errors.New("invalid wal data directory")
