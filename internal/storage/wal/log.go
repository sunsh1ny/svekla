package wal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Log struct {
	options Options
	logger  *zap.Logger

	requests chan appendRequest
	done     chan struct{}
	closed   chan struct{}

	closeOnce sync.Once
	closeErr  error
}

type appendRequest struct {
	record Record
	result chan error
}

type segmentWriter struct {
	file  *os.File
	index int64
	size  int64
}

func Open(options Options, logger *zap.Logger) (*Log, error) {
	if err := options.validate(); err != nil {
		return nil, err
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	if err := os.MkdirAll(options.DataDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("create wal directory: %w", err)
	}

	log := &Log{
		options:  options,
		logger:   logger.Named("wal"),
		requests: make(chan appendRequest),
		done:     make(chan struct{}),
		closed:   make(chan struct{}),
	}

	writer, err := log.openWriter()
	if err != nil {
		return nil, err
	}

	go log.run(writer)

	return log, nil
}

func (l *Log) Append(record Record) error {
	if err := record.Validate(); err != nil {
		return err
	}

	request := appendRequest{
		record: record,
		result: make(chan error, 1),
	}

	select {
	case <-l.done:
		return ErrClosed
	case l.requests <- request:
	}

	select {
	case err := <-request.result:
		return err
	case <-l.done:
		return ErrClosed
	}
}

func (l *Log) Close() error {
	l.closeOnce.Do(func() {
		close(l.done)
		<-l.closed
	})

	return l.closeErr
}

func Replay(dir string, apply func(Record) error) error {
	segments, err := listSegments(dir)
	if err != nil {
		return err
	}

	for _, segment := range segments {
		if err := replaySegment(segment.path, apply); err != nil {
			return err
		}
	}

	return nil
}

func replaySegment(path string, apply func(Record) error) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open wal segment %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)

	for scanner.Scan() {
		record, err := decodeRecord(scanner.Bytes())
		if err != nil {
			return fmt.Errorf("decode wal segment %q: %w", path, err)
		}

		if err := apply(record); err != nil {
			return fmt.Errorf("apply wal segment %q: %w", path, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read wal segment %q: %w", path, err)
	}

	return nil
}

func (l *Log) run(writer *segmentWriter) {
	defer close(l.closed)
	defer func() {
		if writer != nil && writer.file != nil {
			if err := writer.file.Close(); err != nil && l.closeErr == nil {
				l.closeErr = err
			}
		}
	}()

	for {
		select {
		case <-l.done:
			return
		case request := <-l.requests:
			batch := []appendRequest{request}
			l.collectBatch(&batch)

			if err := l.writeBatch(&writer, batch); err != nil {
				l.logger.Error("write wal batch", zap.Error(err))
				l.replyBatch(batch, err)
				l.failUntilClosed(err)
				return
			}

			l.replyBatch(batch, nil)
		}
	}
}

func (l *Log) collectBatch(batch *[]appendRequest) {
	if len(*batch) >= l.options.BatchSize {
		return
	}

	timer := time.NewTimer(l.options.BatchTimeout)
	defer timer.Stop()

	for len(*batch) < l.options.BatchSize {
		select {
		case request := <-l.requests:
			*batch = append(*batch, request)
		case <-timer.C:
			return
		case <-l.done:
			return
		}
	}
}

func (l *Log) writeBatch(writer **segmentWriter, batch []appendRequest) error {
	payload := make([][]byte, 0, len(batch))
	totalSize := 0

	for _, request := range batch {
		data, err := encodeRecord(request.record)
		if err != nil {
			return err
		}

		payload = append(payload, data)
		totalSize += len(data)
	}

	if err := l.rotateIfNeeded(writer, int64(totalSize)); err != nil {
		return err
	}

	for _, data := range payload {
		n, err := (*writer).file.Write(data)
		if err != nil {
			return fmt.Errorf("write wal segment: %w", err)
		}
		if n != len(data) {
			return io.ErrShortWrite
		}
		(*writer).size += int64(n)
	}

	if err := (*writer).file.Sync(); err != nil {
		return fmt.Errorf("fsync wal segment: %w", err)
	}

	return nil
}

func (l *Log) rotateIfNeeded(writer **segmentWriter, incomingSize int64) error {
	if (*writer).size == 0 || (*writer).size+incomingSize <= int64(l.options.SegmentSize) {
		return nil
	}

	nextIndex := (*writer).index + 1
	if err := (*writer).file.Close(); err != nil {
		return fmt.Errorf("close wal segment: %w", err)
	}

	next, err := openSegment(l.options.DataDirectory, nextIndex)
	if err != nil {
		return err
	}

	*writer = next
	return nil
}

func (l *Log) replyBatch(batch []appendRequest, err error) {
	for _, request := range batch {
		request.result <- err
	}
}

func (l *Log) failUntilClosed(err error) {
	for {
		select {
		case <-l.done:
			return
		case request := <-l.requests:
			request.result <- err
		}
	}
}

func (l *Log) openWriter() (*segmentWriter, error) {
	segments, err := listSegments(l.options.DataDirectory)
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return openSegment(l.options.DataDirectory, 1)
	}

	last := segments[len(segments)-1]
	if last.size >= int64(l.options.SegmentSize) {
		return openSegment(l.options.DataDirectory, last.index+1)
	}

	return openExistingSegment(last)
}

func openSegment(dir string, index int64) (*segmentWriter, error) {
	path := segmentPath(dir, index)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create wal segment %q: %w", path, err)
	}

	return &segmentWriter{
		file:  file,
		index: index,
	}, nil
}

func openExistingSegment(segment segmentInfo) (*segmentWriter, error) {
	file, err := os.OpenFile(segment.path, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open wal segment %q: %w", segment.path, err)
	}

	return &segmentWriter{
		file:  file,
		index: segment.index,
		size:  segment.size,
	}, nil
}
