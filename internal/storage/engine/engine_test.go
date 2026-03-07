package engine

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestEngine_Set(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantValue string
		wantFound bool
		wantErr   error
	}{
		{
			name:    "set empty key",
			key:     "",
			value:   "test",
			wantErr: ErrEmptyKey,
		},
		{
			name:      "set success",
			key:       "test",
			value:     "value",
			wantValue: "value",
			wantFound: true,
		},
		{
			name:    "set whitespace key",
			key:     "   ",
			value:   "value",
			wantErr: ErrEmptyKey,
		},
	}

	logger := zap.NewNop()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := NewEngine(logger)

			err := engine.Set(test.key, test.value)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}

			if test.wantErr != nil {
				return
			}

			gotValue, gotFound, err := engine.Get(test.key)
			if err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}

			if gotFound != test.wantFound {
				t.Fatalf("got found=%v, want found=%v", gotFound, test.wantFound)
			}

			if gotValue != test.wantValue {
				t.Fatalf("got value=%q, want value=%q", gotValue, test.wantValue)
			}
		})
	}
}

func TestEngine_SetOverwrite(t *testing.T) {
	logger := zap.NewNop()

	engine := NewEngine(logger)

	if err := engine.Set("key", "value1"); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	if err := engine.Set("key", "value2"); err != nil {
		t.Fatalf("unexpected overwrite error: %v", err)
	}

	got, found, err := engine.Get("key")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if !found {
		t.Fatalf("expected key to be found")
	}

	if got != "value2" {
		t.Fatalf("got value=%q, want value=%q", got, "value2")
	}
}

func TestEngine_Get(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		prepare   bool
		wantValue string
		wantFound bool
		wantErr   error
	}{
		{
			name:    "get empty key",
			key:     "",
			value:   "test",
			wantErr: ErrEmptyKey,
		},
		{
			name:    "get space key",
			key:     "   ",
			value:   "value",
			wantErr: ErrEmptyKey,
		},
		{
			name:    "get tab key",
			key:     "\t",
			value:   "value",
			wantErr: ErrEmptyKey,
		},
		{
			name:      "get success",
			key:       "key",
			value:     "value",
			wantValue: "value",
			prepare:   true,
			wantFound: true,
		},
		{
			name:      "missing valid key",
			key:       "missing",
			wantValue: "",
			wantFound: false,
			wantErr:   nil,
		},
	}

	logger := zap.NewNop()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := NewEngine(logger)

			if test.prepare {
				err := engine.Set(test.key, test.value)
				if err != nil {
					t.Fatalf("unexpected setup error %v", err)
				}
			}

			gotValue, gotFound, err := engine.Get(test.key)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}

			if test.wantErr != nil {
				return
			}

			if gotValue != test.wantValue {
				t.Fatalf("got value=%q, want value=%q", gotValue, test.wantValue)
			}

			if gotFound != test.wantFound {
				t.Fatalf("got found=%v, want found=%v", gotFound, test.wantFound)
			}
		})
	}
}

func TestEngine_Delete(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		prepare bool
		ok      bool
		wantErr error
	}{
		{
			name:    "delete empty key",
			key:     "",
			ok:      false,
			wantErr: ErrEmptyKey,
		},
		{
			name:    "delete success",
			key:     "key",
			value:   "value",
			prepare: true,
			ok:      true,
		},
		{
			name:    "delete whitespace key",
			key:     " ",
			wantErr: ErrEmptyKey,
		},
	}

	logger := zap.NewNop()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := NewEngine(logger)
			if test.prepare {
				err := engine.Set(test.key, test.value)
				if err != nil {
					t.Fatalf("unexpected setup error %v", err)
				}
			}

			ok, err := engine.Delete(test.key)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}

			if test.wantErr != nil {
				return
			}

			if ok != test.ok {
				t.Fatalf("got ok=%v, want ok=%v", ok, test.ok)
			}

			_, gotFound, err := engine.Get(test.key)
			if err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
			if gotFound != false {
				t.Fatalf("got found=%v, want found=%v", gotFound, false)
			}
		})
	}
}

func TestEngine_DeleteMissingKey(t *testing.T) {
	logger := zap.NewNop()
	engine := NewEngine(logger)

	deleted, err := engine.Delete("missing")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deleted {
		t.Fatalf("expected deleted=false for missing key")
	}
}
