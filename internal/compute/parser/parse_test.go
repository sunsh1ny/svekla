package parser

import (
	"errors"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Query
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: ErrEmptyQuery,
		},
		{
			name:    "unknown command",
			input:   "LOL kek",
			wantErr: ErrUnknownCommand,
		},
		{
			name:    "unknown command ~> case sensitive",
			input:   "set key value",
			wantErr: ErrUnknownCommand,
		},
		{
			name:  "positive set",
			input: "SET key value",
			want: Query{
				commandID: SetCommandID,
				arguments: []string{"key", "value"},
			},
		},
		{
			name:  "positive get",
			input: "GET key",
			want: Query{
				commandID: GetCommandID,
				arguments: []string{"key"},
			},
		},
		{
			name:  "positive del",
			input: "DEL key",
			want: Query{
				commandID: DelCommandID,
				arguments: []string{"key"},
			},
		},
		{
			name:    "wrong arg count (more) set",
			input:   "SET key value1 value2",
			wantErr: ErrInvalidArgumentsNumber,
		},
		{
			name:    "wrong arg count (less) set",
			input:   "SET value1",
			wantErr: ErrInvalidArgumentsNumber,
		},
		{
			name:    "wrong arg count (more) get",
			input:   "GET key1 key2",
			wantErr: ErrInvalidArgumentsNumber,
		},
		{
			name:    "wrong arg count (less) get",
			input:   "GET ",
			wantErr: ErrInvalidArgumentsNumber,
		},
		{
			name:    "wrong arg count (more) del",
			input:   "DEL key1 key2",
			wantErr: ErrInvalidArgumentsNumber,
		},
		{
			name:    "wrong arg count (less) del",
			input:   "DEL ",
			wantErr: ErrInvalidArgumentsNumber,
		},
		{
			name:  "spaces",
			input: "SET         key                value",
			want: Query{
				commandID: SetCommandID,
				arguments: []string{"key", "value"},
			},
		},
		{
			name:  "leading & trailing spaces",
			input: "   SET key value   ",
			want: Query{
				commandID: SetCommandID,
				arguments: []string{"key", "value"},
			},
		},
		{
			name:  "tabs",
			input: "SET \tkey\tvalue",
			want: Query{
				commandID: SetCommandID,
				arguments: []string{"key", "value"},
			},
		},
	}

	logger := zap.NewNop()
	p := NewParser(logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := p.Parse(test.input)

			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("expected error %v, got %v", test.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(res, test.want) {
				t.Fatalf("got %#v, want %#v", res, test.want)
			}
		})
	}
}
