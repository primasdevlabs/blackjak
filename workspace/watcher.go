package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type NormalizedEvent struct {
	Path      string
	Type      ChangeType
	Timestamp time.Time
}

type WatcherHandler func(event NormalizedEvent)

// WorkspaceWatcher monitors filesystem changes, normalizes multiple raw write events,
// and reconciles modifications regardless of source (tools, shell, IDE, formatters, git).
type WorkspaceWatcher struct {
	mu           sync.Mutex
	rootPath     string
	handler      WatcherHandler
	pending      map[string]*time.Timer
	lastHashes   map[string]string
	debounceTime time.Duration
	stopCh       chan struct{}
}

func NewWorkspaceWatcher(rootPath string, handler WatcherHandler) *WorkspaceWatcher {
	return &WorkspaceWatcher{
		rootPath:     rootPath,
		handler:      handler,
		pending:      make(map[string]*time.Timer),
		lastHashes:   make(map[string]string),
		debounceTime: 250 * time.Millisecond,
		stopCh:       make(chan struct{}),
	}
}

func (w *WorkspaceWatcher) Start(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-ticker.C:
				w.scanWorkspace()
			}
		}
	}()
}

func (w *WorkspaceWatcher) Stop() {
	close(w.stopCh)
}

func (w *WorkspaceWatcher) scanWorkspace() {
	if w.rootPath == "" {
		return
	}

	_ = filepath.Walk(w.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() && shouldIgnoreDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate file hash to detect external / tool writes
		currentHash := CalculateFileHash(path)
		if currentHash == "" {
			return nil
		}

		w.mu.Lock()
		prevHash, exists := w.lastHashes[path]
		w.lastHashes[path] = currentHash
		w.mu.Unlock()

		if !exists {
			// First scan or created
			w.emitDebounced(path, ChangeCreated)
		} else if prevHash != currentHash {
			// File content modified
			w.emitDebounced(path, ChangeModified)
		}

		return nil
	})
}

func (w *WorkspaceWatcher) emitDebounced(path string, cType ChangeType) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if timer, ok := w.pending[path]; ok {
		timer.Stop()
	}

	w.pending[path] = time.AfterFunc(w.debounceTime, func() {
		w.mu.Lock()
		delete(w.pending, path)
		w.mu.Unlock()

		if w.handler != nil {
			w.handler(NormalizedEvent{
				Path:      path,
				Type:      cType,
				Timestamp: time.Now(),
			})
		}
	})
}

func shouldIgnoreDir(name string) bool {
	ignored := []string{".git", "node_modules", "dist", "bin", "vendor", ".gemini", ".vscode"}
	for _, ign := range ignored {
		if strings.EqualFold(name, ign) {
			return true
		}
	}
	return false
}
