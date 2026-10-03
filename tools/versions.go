package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"blackjak/workspace"
)

// VersionsTool reports current dependency/tooling versions from the workspace.
type VersionsTool struct {
	ws *workspace.Workspace
}

func (t *VersionsTool) Name() string { return "versions" }

func (t *VersionsTool) Description() string {
	return "Inspect current language/tool and module versions in the workspace (go list, package.json engines/deps). Prefer this over guessing from training data."
}

func (t *VersionsTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ecosystem": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"auto", "go", "node"},
				"description": "Which ecosystem to inspect (default auto)",
			},
		},
	}
}

func (t *VersionsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	eco := argString(args, "ecosystem")
	if eco == "" {
		eco = "auto"
	}
	root := t.ws.RootPath
	out := map[string]interface{}{}

	if eco == "auto" || eco == "go" {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			goInfo := map[string]interface{}{}
			if ver, err := runQuick(ctx, root, "go", "version"); err == nil {
				goInfo["go"] = strings.TrimSpace(ver)
			}
			if mod, err := runQuick(ctx, root, "go", "list", "-m"); err == nil {
				goInfo["module"] = strings.TrimSpace(mod)
			}
			if deps, err := runQuick(ctx, root, "go", "list", "-m", "all"); err == nil {
				lines := strings.Split(strings.TrimSpace(deps), "\n")
				if len(lines) > 40 {
					lines = lines[:40]
					goInfo["modulesTruncated"] = true
				}
				goInfo["modules"] = lines
			}
			out["go"] = goInfo
		}
	}

	if eco == "auto" || eco == "node" {
		pkgPath := filepath.Join(root, "package.json")
		if data, err := os.ReadFile(pkgPath); err == nil {
			var pkg map[string]interface{}
			if json.Unmarshal(data, &pkg) == nil {
				nodeInfo := map[string]interface{}{}
				if eng, ok := pkg["engines"]; ok {
					nodeInfo["engines"] = eng
				}
				if deps, ok := pkg["dependencies"]; ok {
					nodeInfo["dependencies"] = deps
				}
				if deps, ok := pkg["devDependencies"]; ok {
					nodeInfo["devDependencies"] = deps
				}
				out["node"] = nodeInfo
			}
		}
		// Also check extension webview package.json common in this monorepo.
		alt := filepath.Join(root, "extension", "webview", "react", "package.json")
		if data, err := os.ReadFile(alt); err == nil {
			var pkg map[string]interface{}
			if json.Unmarshal(data, &pkg) == nil {
				out["nodeWebview"] = map[string]interface{}{
					"path":         "extension/webview/react/package.json",
					"dependencies": pkg["dependencies"],
				}
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no go.mod or package.json found for version inspection")
	}
	return out, nil
}

func runQuick(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, name, args...)
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
