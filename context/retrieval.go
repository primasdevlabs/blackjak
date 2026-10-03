package context

import (
	"os"
	"path/filepath"
	"strings"
)

// Retriever selects relevant items to inject into context.
type Retriever struct {
	Root  string
	Index SearchIndex
}

// SearchIndex is satisfied by workspace.FileIndex.
type SearchIndex interface {
	Search(query string, limit int) []string
}

// NewRetriever initializes a new Retriever.
func NewRetriever(root string, index SearchIndex) *Retriever {
	return &Retriever{Root: root, Index: index}
}

// Retrieve returns path snippets relevant to the query.
func (r *Retriever) Retrieve(query string, limit int) []string {
	if limit <= 0 {
		limit = 8
	}
	var paths []string
	if r.Index != nil {
		paths = r.Index.Search(query, limit*2)
	}
	var out []string
	for _, p := range paths {
		if len(out) >= limit {
			break
		}
		snippet := r.readSnippet(p, 40)
		if snippet == "" {
			out = append(out, p)
			continue
		}
		out = append(out, p+"\n"+snippet)
	}
	return out
}

func (r *Retriever) readSnippet(rel string, maxLines int) string {
	full := filepath.Join(r.Root, filepath.FromSlash(rel))
	data, err := os.ReadFile(full)
	if err != nil {
		return ""
	}
	// Skip large/binary-ish files
	if len(data) > 64*1024 {
		data = data[:64*1024]
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}
