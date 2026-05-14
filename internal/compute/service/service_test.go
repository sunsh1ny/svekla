package service

import (
	"errors"
	"testing"

	"svekla/internal/compute/parser"
	"svekla/internal/storage/engine"

	"go.uber.org/zap"
)

func TestCommandServiceExecute(t *testing.T) {
	commandService := newTestCommandService()

	tests := []struct {
		name    string
		query   parser.Query
		want    string
		wantErr error
	}{
		{
			name:  "set",
			query: parser.NewQuery(parser.SetCommandID, []string{"key", "value"}),
			want:  ResultOK,
		},
		{
			name:  "get",
			query: parser.NewQuery(parser.GetCommandID, []string{"key"}),
			want:  "value",
		},
		{
			name:  "delete",
			query: parser.NewQuery(parser.DelCommandID, []string{"key"}),
			want:  ResultOK,
		},
		{
			name:  "get missing",
			query: parser.NewQuery(parser.GetCommandID, []string{"key"}),
			want:  ResultNotFound,
		},
		{
			name:    "unknown command",
			query:   parser.NewQuery(parser.UnknownCommandID, nil),
			wantErr: ErrUnknownCommand,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := commandService.Execute(test.query)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}

			if test.wantErr != nil {
				return
			}

			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestCommandServiceExecute_InvalidArguments(t *testing.T) {
	commandService := newTestCommandService()

	tests := []struct {
		name  string
		query parser.Query
	}{
		{
			name:  "set without value",
			query: parser.NewQuery(parser.SetCommandID, []string{"key"}),
		},
		{
			name:  "get without key",
			query: parser.NewQuery(parser.GetCommandID, nil),
		},
		{
			name:  "delete without key",
			query: parser.NewQuery(parser.DelCommandID, nil),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := commandService.Execute(test.query)

			if !errors.Is(err, ErrInvalidCommandArguments) {
				t.Fatalf("expected error %v, got %v", ErrInvalidCommandArguments, err)
			}
		})
	}
}

func newTestCommandService() *CommandService {
	logger := zap.NewNop()
	store := engine.NewEngine(logger)

	return NewCommandService(store)
}
