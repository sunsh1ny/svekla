package wal

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"svekla/internal/storage/engine"

	"go.uber.org/zap"
)

func TestLogAppendAndReplay(t *testing.T) {
	dir := t.TempDir()
	log := openTestLog(t, dir, Options{
		BatchSize:    1,
		BatchTimeout: time.Second,
		SegmentSize:  1024,
	})

	if err := log.Append(NewSetRecord("key", "value")); err != nil {
		t.Fatalf("append set: %v", err)
	}
	if err := log.Append(NewDeleteRecord("missing")); err != nil {
		t.Fatalf("append delete: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("close wal: %v", err)
	}

	var got []Record
	if err := Replay(dir, func(record Record) error {
		got = append(got, record)
		return nil
	}); err != nil {
		t.Fatalf("replay wal: %v", err)
	}

	want := []Record{
		NewSetRecord("key", "value"),
		NewDeleteRecord("missing"),
	}

	if len(got) != len(want) {
		t.Fatalf("got %d records, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestLogAppendFlushesByBatchSize(t *testing.T) {
	dir := t.TempDir()
	log := openTestLog(t, dir, Options{
		BatchSize:    2,
		BatchTimeout: time.Second,
		SegmentSize:  1024,
	})
	defer log.Close()

	firstDone := make(chan error, 1)
	go func() {
		firstDone <- log.Append(NewSetRecord("key1", "value1"))
	}()

	select {
	case err := <-firstDone:
		t.Fatalf("first append flushed before batch was full: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	secondDone := make(chan error, 1)
	go func() {
		secondDone <- log.Append(NewSetRecord("key2", "value2"))
	}()

	assertAppendDone(t, firstDone)
	assertAppendDone(t, secondDone)
}

func TestLogAppendFlushesByTimeout(t *testing.T) {
	dir := t.TempDir()
	log := openTestLog(t, dir, Options{
		BatchSize:    10,
		BatchTimeout: 20 * time.Millisecond,
		SegmentSize:  1024,
	})
	defer log.Close()

	done := make(chan error, 1)
	go func() {
		done <- log.Append(NewSetRecord("key", "value"))
	}()

	assertAppendDone(t, done)
}

func TestLogRotatesSegments(t *testing.T) {
	dir := t.TempDir()
	log := openTestLog(t, dir, Options{
		BatchSize:    1,
		BatchTimeout: time.Second,
		SegmentSize:  70,
	})

	for i := 0; i < 5; i++ {
		if err := log.Append(NewSetRecord(string(rune('a'+i)), "value")); err != nil {
			t.Fatalf("append record %d: %v", i, err)
		}
	}

	if err := log.Close(); err != nil {
		t.Fatalf("close wal: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(dir, segmentFilePrefix+"*"+segmentFileSuffix))
	if err != nil {
		t.Fatalf("glob wal segments: %v", err)
	}

	if len(matches) < 2 {
		t.Fatalf("got %d segment(s), want at least 2", len(matches))
	}
}

func TestDurableStoreRecoversState(t *testing.T) {
	dir := t.TempDir()
	logger := zap.NewNop()

	base := engine.NewEngine(logger)
	log := openTestLog(t, dir, Options{
		BatchSize:    1,
		BatchTimeout: time.Second,
		SegmentSize:  1024,
	})

	store, err := NewDurableStore(base, log)
	if err != nil {
		t.Fatalf("new durable store: %v", err)
	}

	if err := store.Set("alive", "yes"); err != nil {
		t.Fatalf("set alive: %v", err)
	}
	if err := store.Set("deleted", "gone"); err != nil {
		t.Fatalf("set deleted: %v", err)
	}
	if ok, err := store.Delete("deleted"); err != nil || !ok {
		t.Fatalf("delete deleted: ok=%v err=%v", ok, err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("close wal: %v", err)
	}

	recovered := engine.NewEngine(logger)
	if err := Recover(recovered, dir); err != nil {
		t.Fatalf("recover wal: %v", err)
	}

	value, ok, err := recovered.Get("alive")
	if err != nil {
		t.Fatalf("get alive: %v", err)
	}
	if !ok || value != "yes" {
		t.Fatalf("alive = %q, %v; want %q, true", value, ok, "yes")
	}

	_, ok, err = recovered.Get("deleted")
	if err != nil {
		t.Fatalf("get deleted: %v", err)
	}
	if ok {
		t.Fatalf("deleted key was recovered")
	}
}

func TestLogAppendAfterClose(t *testing.T) {
	dir := t.TempDir()
	log := openTestLog(t, dir, Options{
		BatchSize:    1,
		BatchTimeout: time.Second,
		SegmentSize:  1024,
	})

	if err := log.Close(); err != nil {
		t.Fatalf("close wal: %v", err)
	}

	err := log.Append(NewSetRecord("key", "value"))
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("got error %v, want %v", err, ErrClosed)
	}
}

func openTestLog(t *testing.T, dir string, options Options) *Log {
	t.Helper()

	options.DataDirectory = dir
	log, err := Open(options, zap.NewNop())
	if err != nil {
		t.Fatalf("open wal: %v", err)
	}

	return log
}

func assertAppendDone(t *testing.T, done <-chan error) {
	t.Helper()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("append error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("append did not finish in time")
	}
}
