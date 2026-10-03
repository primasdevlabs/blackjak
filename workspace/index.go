package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// IndexStats reports workspace index state.
type IndexStats struct {
	FileCount  int       `json:"fileCount"`
	LastIndexed time.Time `json:"lastIndexed"`
	Status     string    `json:"status"`
	Root       string    `json:"root"`
}

// RepoSnapshot is a lightweight progressive model of the repository.
type RepoSnapshot struct {
	Languages  []string `json:"languages"`
	Entrypoints []string `json:"entrypoints"`
	BuildFiles []string `json:"buildFiles"`
	FileCount  int      `json:"fileCount"`
}

// FileIndex is a simple path/keyword index for retrieval.
type FileIndex struct {
	mu       sync.RWMutex
	root     string
	files    []string
	updated  time.Time
	status   string
	snapshot RepoSnapshot
}

func NewFileIndex(root string) *FileIndex {
	fi := &FileIndex{root: root, status: "idle"}
	go fi.Rebuild()
	return fi
}

func (fi *FileIndex) Stats() IndexStats {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	return IndexStats{
		FileCount:   len(fi.files),
		LastIndexed: fi.updated,
		Status:      fi.status,
		Root:        fi.root,
	}
}

func (fi *FileIndex) Rebuild() {
	fi.mu.Lock()
	fi.status = "indexing"
	fi.mu.Unlock()

	skip := map[string]bool{
		".git": true, "node_modules": true, "dist": true, "out": true,
		"bin": true, ".blackjak": true,
	}
	var files []string
	_ = filepath.WalkDir(fi.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skip[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(fi.root, path)
		if err != nil {
			return nil
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})

	snap := buildRepoSnapshot(files)

	fi.mu.Lock()
	fi.files = files
	fi.snapshot = snap
	fi.updated = time.Now()
	fi.status = "ready"
	fi.mu.Unlock()
}

// Snapshot returns a copy of the lightweight repo model.
func (fi *FileIndex) Snapshot() RepoSnapshot {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	out := fi.snapshot
	out.Languages = append([]string{}, fi.snapshot.Languages...)
	out.Entrypoints = append([]string{}, fi.snapshot.Entrypoints...)
	out.BuildFiles = append([]string{}, fi.snapshot.BuildFiles...)
	return out
}

func buildRepoSnapshot(files []string) RepoSnapshot {
	langCount := map[string]int{}
	var entrypoints, buildFiles []string
	entryHints := map[string]bool{
		"main.go": true, "cmd/agent/main.go": true, "package.json": true,
		"Dockerfile": true, "docker-compose.yml": true, "docker-compose.yaml": true,
		"index.ts": true, "index.js": true, "src/main.tsx": true, "src/main.ts": true,
		"app.py": true, "manage.py": true,
	}
	buildHints := map[string]bool{
		"go.mod": true, "go.sum": true, "package.json": true, "package-lock.json": true,
		"Cargo.toml": true, "pom.xml": true, "build.gradle": true, "Makefile": true,
		"tsconfig.json": true, "pyproject.toml": true, "requirements.txt": true,
		"CMakeLists.txt": true,
	}
	extLang := map[string]string{
		".go": "Go", ".ts": "TypeScript", ".tsx": "TypeScript", ".js": "JavaScript",
		".jsx": "JavaScript", ".py": "Python", ".rs": "Rust", ".java": "Java",
		".cs": "C#", ".rb": "Ruby", ".php": "PHP", ".css": "CSS", ".html": "HTML",
		".md": "Markdown", ".sql": "SQL", ".yml": "YAML", ".yaml": "YAML",
	}
	for _, f := range files {
		base := filepath.Base(f)
		ext := strings.ToLower(filepath.Ext(f))
		if lang, ok := extLang[ext]; ok {
			langCount[lang]++
		}
		if entryHints[f] || entryHints[base] || strings.HasSuffix(f, "/main.go") {
			entrypoints = append(entrypoints, f)
		}
		if buildHints[f] || buildHints[base] {
			buildFiles = append(buildFiles, f)
		}
	}
	type kv struct {
		k string
		n int
	}
	var langs []kv
	for k, n := range langCount {
		langs = append(langs, kv{k, n})
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].n > langs[j].n })
	var languages []string
	for i, l := range langs {
		if i >= 8 {
			break
		}
		languages = append(languages, l.k)
	}
	if len(entrypoints) > 12 {
		entrypoints = entrypoints[:12]
	}
	if len(buildFiles) > 12 {
		buildFiles = buildFiles[:12]
	}
	return RepoSnapshot{
		Languages:   languages,
		Entrypoints: entrypoints,
		BuildFiles:  buildFiles,
		FileCount:   len(files),
	}
}

func (fi *FileIndex) Search(query string, limit int) []string {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	if limit <= 0 {
		limit = 20
	}
	q := strings.ToLower(query)
	var out []string
	for _, f := range fi.files {
		if q == "" || strings.Contains(strings.ToLower(f), q) {
			out = append(out, f)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}
