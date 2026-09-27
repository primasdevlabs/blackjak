package workspace

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ReferenceType defines whether a reference targets a file or a folder.
type ReferenceType string

const (
	RefTypeFile   ReferenceType = "file"
	RefTypeFolder ReferenceType = "folder"
)

// AttachmentType defines the attachment item type.
type AttachmentType string

const (
	AttachTypeFile   AttachmentType = "file"
	AttachTypeFolder AttachmentType = "folder"
)

// Attachment represents a file or folder explicitly attached to a task request.
type Attachment struct {
	ID   string         `json:"id"`
	Type AttachmentType `json:"type"`
	Path string         `json:"path"`
	Name string         `json:"name"`
}

// WorkspaceReference represents a structured workspace reference resolved from user prompt.
type WorkspaceReference struct {
	Type ReferenceType `json:"type"`
	Path string        `json:"path"`
	Line int           `json:"line,omitempty"`
	Raw  string        `json:"raw"`
}

var refRegex = regexp.MustCompile(`@(?:(file|folder):)?([^\s\t\r\n,;]+)`)

// ParseWorkspaceReferences scans text for @path syntax and resolves valid workspace references.
func (w *Workspace) ParseWorkspaceReferences(prompt string) ([]WorkspaceReference, string) {
	matches := refRegex.FindAllStringSubmatch(prompt, -1)
	if len(matches) == 0 {
		return nil, prompt
	}

	refs := make([]WorkspaceReference, 0, len(matches))
	cleanedPrompt := prompt

	for _, m := range matches {
		raw := m[0]
		typePrefix := strings.ToLower(m[1])
		targetPath := m[2]

		// Extract line number if specified as path:line (e.g. middleware.go:84)
		line := 0
		if idx := strings.LastIndex(targetPath, ":"); idx != -1 {
			if l, err := strconv.Atoi(targetPath[idx+1:]); err == nil {
				line = l
				targetPath = targetPath[:idx]
			}
		}

		resolvedPath, err := w.ResolvePath(targetPath)
		if err != nil {
			// Skip paths violating workspace isolation boundary
			continue
		}

		relPath, _ := filepath.Rel(w.RootPath, resolvedPath)
		relPath = filepath.ToSlash(relPath)

		refType := RefTypeFile
		if typePrefix == "folder" || strings.HasSuffix(relPath, "/") || filepath.Ext(relPath) == "" {
			refType = RefTypeFolder
		}
		if typePrefix == "file" {
			refType = RefTypeFile
		}

		refs = append(refs, WorkspaceReference{
			Type: refType,
			Path: relPath,
			Line: line,
			Raw:  raw,
		})
	}

	return refs, cleanedPrompt
}

// ValidateAttachments verifies that attached file/folder paths belong to the workspace boundary.
func (w *Workspace) ValidateAttachments(attachments []Attachment) ([]Attachment, error) {
	valid := make([]Attachment, 0, len(attachments))
	for _, att := range attachments {
		resolved, err := w.ResolvePath(att.Path)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(w.RootPath, resolved)
		att.Path = filepath.ToSlash(rel)
		valid = append(valid, att)
	}
	return valid, nil
}
