package memory

// WorkingMemory maintains short-term conversational and step context.
type WorkingMemory struct {
	items map[string]interface{}
}

// NewWorkingMemory creates a new WorkingMemory instance.
func NewWorkingMemory() *WorkingMemory {
	return &WorkingMemory{
		items: make(map[string]interface{}),
	}
}
