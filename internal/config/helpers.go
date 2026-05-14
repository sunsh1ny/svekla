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
	return parseByteSize(n.MaxMessageSize)
}

func parseByteSize(input string) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(input))
	if raw == "" {
		return 0, ErrInvalidMaxMessageSize
	}

	for _, u := range byteSizeUnits {
		if strings.HasSuffix(raw, u.suffix) {
			return parseByteSizeValue(raw, u)
		}
	}

	return 0, ErrInvalidMaxMessageSize
}

func parseByteSizeValue(raw string, unit byteSizeUnit) (int, error) {
	numberPart := strings.TrimSpace(strings.TrimSuffix(raw, unit.suffix))
	if numberPart == "" {
		return 0, ErrInvalidMaxMessageSize
	}

	value, err := strconv.Atoi(numberPart)
	if err != nil || value <= 0 {
		return 0, ErrInvalidMaxMessageSize
	}

	return value * unit.multiplier, nil
}
