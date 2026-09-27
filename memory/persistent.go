package memory

// PersistentMemory manages long-term memory across sessions.
type PersistentMemory struct{}

// NewPersistentMemory initializes PersistentMemory.
func NewPersistentMemory() *PersistentMemory {
	return &PersistentMemory{}
}
