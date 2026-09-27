package memory

import (
	"context"
	"sort"
	"sync"
	"time"
)

// WorkingItem is a single entry in working memory.
type WorkingItem struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// WorkingMemory maintains short-term conversational and step context for a
// single run. It is thread-safe and discarded when the run ends.
type WorkingMemory struct {
	mu    sync.RWMutex
	items map[string]*WorkingItem
	order []string // insertion order for deterministic Keys()
}

// NewWorkingMemory creates a new WorkingMemory instance.
func NewWorkingMemory() *WorkingMemory {
	return &WorkingMemory{
		items: make(map[string]*WorkingItem),
	}
}

// Get retrieves a value by key.
func (m *WorkingMemory) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.items[key]
	if !ok {
		return nil, nil
	}
	return item.Value, nil
}

// Set stores a value under key, preserving insertion order.
func (m *WorkingMemory) Set(ctx context.Context, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	if item, ok := m.items[key]; ok {
		item.Value = value
		item.UpdatedAt = now
		return nil
	}
	m.items[key] = &WorkingItem{Key: key, Value: value, CreatedAt: now, UpdatedAt: now}
	m.order = append(m.order, key)
	return nil
}

// Delete removes a key.
func (m *WorkingMemory) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
	for i, k := range m.order {
		if k == key {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
	return nil
}

// Keys returns all keys in insertion order.
func (m *WorkingMemory) Keys(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.order))
	copy(out, m.order)
	return out, nil
}

// Items returns all items sorted by key for prompt injection.
func (m *WorkingMemory) Items() []*WorkingItem {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*WorkingItem, 0, len(m.items))
	for _, item := range m.items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
