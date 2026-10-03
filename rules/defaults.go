package rules

import (
	"embed"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed defaults/*.md
var builtinRulesFS embed.FS

// BuiltinRules returns the engineering packs shipped inside the agent binary.
// These apply on every fresh install even when the workspace has no .blackjak/rules.
func BuiltinRules() []Rule {
	entries, err := fs.ReadDir(builtinRulesFS, "defaults")
	if err != nil {
		return nil
	}
	var out []Rule
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		data, err := builtinRulesFS.ReadFile(path.Join("defaults", e.Name()))
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(e.Name(), path.Ext(e.Name()))
		out = append(out, Rule{
			ID:      "builtin:" + name,
			Name:    name,
			Body:    string(data),
			Source:  "builtin",
			Enabled: true,
			Path:    "builtin://" + name,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
