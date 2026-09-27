package context

// Retriever selects relevant items to inject into context.
type Retriever struct{}

// NewRetriever initializes a new Retriever.
func NewRetriever() *Retriever {
	return &Retriever{}
}
