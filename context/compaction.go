package context

// Compactor handles context window compaction and token reduction.
type Compactor struct{}

// NewCompactor creates a new Compactor instance.
func NewCompactor() *Compactor {
	return &Compactor{}
}
