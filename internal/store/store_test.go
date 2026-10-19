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
