package config

import (
	"strconv"
	"strings"
)

type byteSizeUnit struct {
	suffix     string
	multiplier int
}

var byteSizeUnits = []byteSizeUnit{
	{suffix: "mb", multiplier: 1024 * 1024},
	{suffix: "kb", multiplier: 1024},
	{suffix: "b", multiplier: 1},
}

func (n NetworkConfig) ParsedMaxMessageSizeBytes() (int, error) {
	return parseByteSize(n.MaxMessageSize, ErrInvalidMaxMessageSize)
}

func (w WALConfig) ParsedSegmentSizeBytes() (int, error) {
	return parseByteSize(w.SegmentSize, ErrInvalidWALSegmentSize)
}

func parseByteSize(input string, invalidErr error) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(input))
	if raw == "" {
		return 0, invalidErr
	}

	for _, u := range byteSizeUnits {
		if strings.HasSuffix(raw, u.suffix) {
			return parseByteSizeValue(raw, u, invalidErr)
		}
	}

	return 0, invalidErr
}

func parseByteSizeValue(raw string, unit byteSizeUnit, invalidErr error) (int, error) {
	numberPart := strings.TrimSpace(strings.TrimSuffix(raw, unit.suffix))
	if numberPart == "" {
		return 0, invalidErr
	}

	value, err := strconv.Atoi(numberPart)
	if err != nil || value <= 0 {
		return 0, invalidErr
	}

	return value * unit.multiplier, nil
}
