package tools

import (
	"context"
)

// FilesystemTool enables reading and writing files.
type FilesystemTool struct{}

// NewFilesystemTool initializes a new FilesystemTool.
func NewFilesystemTool() *FilesystemTool {
	return &FilesystemTool{}
}

func (f *FilesystemTool) Name() string { return "filesystem" }

func (f *FilesystemTool) Description() string {
	return "Read and write files within workspace boundary."
}

func (f *FilesystemTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return nil, nil
}
