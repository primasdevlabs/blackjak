package context

// Builder constructs prompts and context payloads for LLM requests.
type Builder struct{}

// NewBuilder creates a new context Builder.
func NewBuilder() *Builder {
	return &Builder{}
}
