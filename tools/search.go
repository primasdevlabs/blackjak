package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"blackjak/workspace"
)

// SearchTool provides regex code search across workspace files.
type SearchTool struct {
	ws *workspace.Workspace
}

type searchMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

func (s *SearchTool) Name() string { return "search" }

func (s *SearchTool) Description() string {
	return "Search file contents with a regular expression. Returns matching lines with file and line numbers. Skips .git, node_modules, vendor, dist."
}

func (s *SearchTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Regular expression to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Subdirectory to search (default: workspace root)",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "File glob filter, e.g. '*.go'",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Max matches (default 100)",
			},
		},
		"required": []string{"pattern"},
	}
}

func (s *SearchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pattern := argString(args, "pattern")
	if pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	root := s.ws.RootPath
	if p := argString(args, "path"); p != "" {
		root, err = s.ws.ResolvePath(p)
		if err != nil {
			return nil, err
		}
	}
	include := argString(args, "include")
	maxResults := argInt(args, "max_results", 100)

	var matches []searchMatch
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if len(matches) >= maxResults {
			return filepath.SkipAll
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if include != "" {
			ok, _ := filepath.Match(include, info.Name())
			if !ok {
				return nil
			}
		}
		if info.Size() > 2*1024*1024 {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		lineNo := 0
		for scanner.Scan() && len(matches) < maxResults {
			lineNo++
			line := scanner.Text()
			if re.MatchString(line) {
				rel, _ := filepath.Rel(s.ws.RootPath, path)
				content := strings.TrimSpace(line)
				if len(content) > 200 {
					content = content[:200] + "…"
				}
				matches = append(matches, searchMatch{
					File:    filepath.ToSlash(rel),
					Line:    lineNo,
					Content: content,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"matches": matches, "count": len(matches)}, nil
}
