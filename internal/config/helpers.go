package config

import (
	"strconv"
	"strings"
)

func (n NetworkConfig) ParsedMaxMessageSizeBytes() (int, error) {
	raw := strings.TrimSpace(strings.ToLower(n.MaxMessageSize))
	if raw == "" {
		return 0, ErrInvalidMaxMessageSize
	}

	type unit struct {
		suffix     string
		multiplier int
	}

	units := []unit{
		{suffix: "mb", multiplier: 1024 * 1024},
		{suffix: "kb", multiplier: 1024},
		{suffix: "b", multiplier: 1},
	}

	for _, u := range units {
		if strings.HasSuffix(raw, u.suffix) {
			numberPart := strings.TrimSpace(strings.TrimSuffix(raw, u.suffix))
			if numberPart == "" {
				return 0, ErrInvalidMaxMessageSize
			}

			value, err := strconv.Atoi(numberPart)
			if err != nil {
				return 0, ErrInvalidMaxMessageSize
			}

			if value <= 0 {
				return 0, ErrInvalidMaxMessageSize
			}

			return value * u.multiplier, nil
		}
	}

	return 0, ErrInvalidMaxMessageSize
}
