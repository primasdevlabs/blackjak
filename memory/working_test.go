package memory

import (
	"context"
	"testing"
)

func TestWorkingMemory_ImplementsStore(t *testing.T) {
	var _ Store = NewWorkingMemory()
}

func TestWorkingMemory_SetGetDelete(t *testing.T) {
	ctx := context.Background()
	m := NewWorkingMemory()

	if err := m.Set(ctx, "goal", "fix the parser"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	v, err := m.Get(ctx, "goal")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v != "fix the parser" {
		t.Errorf("Get returned %v", v)
	}

	if err := m.Delete(ctx, "goal"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	v, _ = m.Get(ctx, "goal")
	if v != nil {
		t.Errorf("expected nil after delete, got %v", v)
	}
}

func TestWorkingMemory_KeysInsertionOrder(t *testing.T) {
	ctx := context.Background()
	m := NewWorkingMemory()

	for _, k := range []string{"c", "a", "b"} {
		if err := m.Set(ctx, k, k); err != nil {
			t.Fatalf("Set %s: %v", k, err)
		}
	}

	keys, err := m.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	want := []string{"c", "a", "b"}
	if len(keys) != len(want) {
		t.Fatalf("expected %d keys, got %v", len(want), keys)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Errorf("key %d: expected %q, got %q", i, want[i], keys[i])
		}
	}

	// Overwriting an existing key must not change order.
	if err := m.Set(ctx, "a", "updated"); err != nil {
		t.Fatalf("Set update: %v", err)
	}
	keys, _ = m.Keys(ctx)
	if len(keys) != 3 || keys[0] != "c" || keys[2] != "b" {
		t.Errorf("order changed after update: %v", keys)
	}
}

func TestWorkingMemory_ItemsSortedByKey(t *testing.T) {
	ctx := context.Background()
	m := NewWorkingMemory()

	for _, k := range []string{"z", "a", "m"} {
		_ = m.Set(ctx, k, k)
	}

	items := m.Items()
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	for i, want := range []string{"a", "m", "z"} {
		if items[i].Key != want {
			t.Errorf("item %d: expected %q, got %q", i, want, items[i].Key)
		}
		if items[i].CreatedAt.IsZero() || items[i].UpdatedAt.IsZero() {
			t.Errorf("item %q missing timestamps", items[i].Key)
		}
	}
}
