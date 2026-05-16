package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	segmentFilePrefix = "wal_"
	segmentFileSuffix = ".log"
)

type segmentInfo struct {
	index int64
	path  string
	size  int64
}

func segmentPath(dir string, index int64) string {
	return filepath.Join(dir, fmt.Sprintf("%s%020d%s", segmentFilePrefix, index, segmentFileSuffix))
}

func listSegments(dir string) ([]segmentInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read wal directory: %w", err)
	}

	segments := make([]segmentInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		index, ok := parseSegmentIndex(entry.Name())
		if !ok {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat wal segment %q: %w", entry.Name(), err)
		}

		segments = append(segments, segmentInfo{
			index: index,
			path:  filepath.Join(dir, entry.Name()),
			size:  info.Size(),
		})
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].index < segments[j].index
	})

	return segments, nil
}

func parseSegmentIndex(name string) (int64, bool) {
	if !strings.HasPrefix(name, segmentFilePrefix) || !strings.HasSuffix(name, segmentFileSuffix) {
		return 0, false
	}

	raw := strings.TrimSuffix(strings.TrimPrefix(name, segmentFilePrefix), segmentFileSuffix)
	index, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}

	return index, true
}
