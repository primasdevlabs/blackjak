package llm

// Tool defines a tool or function schema passed to the LLM interface.
type Tool struct {
	Name        string
	Description string
	Parameters  interface{}
}
