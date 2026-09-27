package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// PersistentEntry is a durable memory record.
type PersistentEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// PersistentMemory manages long-term memory across sessions, backed by a
// JSON file inside the workspace (e.g. .blackjak/memory.json).
type PersistentMemory struct {
	mu    sync.RWMutex
	path  string
	items map[string]*PersistentEntry
}

// NewPersistentMemory initializes a PersistentMemory backed by path.
// The file is created lazily on first write; reads of a missing file yield
// empty memory.
func NewPersistentMemory(path string) *PersistentMemory {
	m := &PersistentMemory{
		path:  path,
		items: make(map[string]*PersistentEntry),
	}
	_ = m.load()
	return m
}

func (m *PersistentMemory) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err // missing file is fine — start empty
	}
	var items map[string]*PersistentEntry
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt memory file %s: %w", m.path, err)
	}
	m.items = items
	return nil
}

func (m *PersistentMemory) flushLocked() error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m.items, "", "  ")
	if err != nil {
		return err
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, m.path)
}

// Get retrieves a durable value.
func (m *PersistentMemory) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if e, ok := m.items[key]; ok {
		return e.Value, nil
	}
	return nil, nil
}

// Set stores a durable value and flushes to disk.
func (m *PersistentMemory) Set(ctx context.Context, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = &PersistentEntry{Key: key, Value: value, UpdatedAt: time.Now()}
	return m.flushLocked()
}

// Delete removes a durable value.
func (m *PersistentMemory) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
	return m.flushLocked()
}

// Keys lists all durable keys.
func (m *PersistentMemory) Keys(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.items))
	for k := range m.items {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

// Entries returns all entries sorted by key for prompt injection.
func (m *PersistentMemory) Entries() []*PersistentEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*PersistentEntry, 0, len(m.items))
	for _, e := range m.items {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
