package store

import "testing"

func TestSetAndGetAndDelete(t *testing.T) {
	store := NewStore()

	store.Set("name", "Magich")

	got, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key: name to exist")
	}

	if got != "Magich" {
		t.Fatalf("got %s, want %s", got, "Magich")
	}

	ok := store.Delete("name")
	if !ok {
		t.Fatal("Could not delete the value with key: name")
	}
}

func TestIncrementOptions(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		exists     bool
		want       int64
		wantErr    bool
		wantStored string
	}{
		{
			name:       "missing key",
			want:       1,
			wantStored: "1",
		},
		{
			name:       "existing integer",
			initial:    "41",
			exists:     true,
			want:       42,
			wantStored: "42",
		},
		{
			name:       "negative integer",
			initial:    "-1",
			exists:     true,
			want:       0,
			wantStored: "0",
		},
		{
			name:       "invalid integer",
			initial:    "hello",
			exists:     true,
			wantErr:    true,
			wantStored: "hello",
		},
		{
			name:       "empty string",
			initial:    "",
			exists:     true,
			wantErr:    true,
			wantStored: "",
		},
		{
			name:       "increment overflow",
			initial:    "9223372036854775807",
			exists:     true,
			wantErr:    true,
			wantStored: "9223372036854775807",
		},
		{
			name:       "integer out of range",
			initial:    "9223372036854775808",
			exists:     true,
			wantErr:    true,
			wantStored: "9223372036854775808",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange: create a fresh store for this case.
			s := NewStore()
			if test.exists {
				s.Set("counter", test.initial)
			}

			got, err := s.Incr("counter")

			// Assert: check whether an error was expected.
			if (err != nil) != test.wantErr {
				t.Fatalf("Incr() error = %v, wantErr = %v", err, test.wantErr)
			}

			// Check the returned number on success.
			if !test.wantErr && got != test.want {
				t.Errorf("Incr() = %d, want %d", got, test.want)
			}

			// Check stored state, including after errors.
			stored, exists := s.Get("counter")
			if !exists {
				t.Fatal("expected counter key to exist")
			}
			if stored != test.wantStored {
				t.Errorf("stored value = %q, want %q", stored, test.wantStored)
			}
		})
	}

}

func TestIncr(t *testing.T) {
	s := NewStore()

	s.Incr("counter")
	s.Incr("counter")
	s.Incr("counter")

	val, _ := s.Get("counter")

	if val != "3" {
		t.Fatal("Counter must be 3")
	}
}
