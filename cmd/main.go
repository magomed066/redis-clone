package main

import (
	"github.com/magomed066/redis-clone/internal/store"
)

func main() {
	store := store.NewStore()

	store.Set("name", "Magich")

	// got, exists := store.Get("name")

	// if !exists {
	// 	panic("expected key to exist")
	// }

	// if got != "Magich" {
	// 	t.Fatalf("got %q, want %q", got, "Magich")
	// }
}
