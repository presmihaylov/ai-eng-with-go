package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema anthropic.ToolInputSchemaParam
	Function    func(input json.RawMessage) (string, error)
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema:"required,description=Path to the file to read"`
}

type ListFilesInput struct {
	Path string `json:"path" jsonschema:"required,description=Path to the directory to list"`
}

type EditFileInput struct {
	Path       string `json:"path" jsonschema:"required,description=Path to the file to edit"`
	OldContent string `json:"old_content" jsonschema:"required,description=Exact content to replace"`
	NewContent string `json:"new_content" jsonschema:"required,description=New content to replace with"`
}

type CreateFileInput struct {
	Path    string `json:"path" jsonschema:"required,description=Path to the new file to create"`
	Content string `json:"content" jsonschema:"required,description=Content to write to the new file"`
}

func generateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

func (a *Agent) setupTools() {
	a.tools = []ToolDefinition{
		{
			Name:        "read_file",
			Description: "Read the contents of a file",
			InputSchema: generateSchema[ReadFileInput](),
			Function:    a.readFile,
		},
		{
			Name:        "list_files",
			Description: "List files and directories in a given path",
			InputSchema: generateSchema[ListFilesInput](),
			Function:    a.listFiles,
		},
		{
			Name:        "edit_file",
			Description: "Edit a file by replacing old content with new content",
			InputSchema: generateSchema[EditFileInput](),
			Function:    a.editFile,
		},
		{
			Name:        "create_file",
			Description: "Create a new file with specified content",
			InputSchema: generateSchema[CreateFileInput](),
			Function:    a.createFile,
		},
	}
}

func (a *Agent) readFile(input json.RawMessage) (string, error) {
	var params ReadFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Check if file is likely binary by extension or lack of extension
	if isBinaryFile(params.Path) {
		return "", fmt.Errorf("cannot read binary file: %s", params.Path)
	}

	content, err := os.ReadFile(params.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", params.Path, err)
	}

	// Additional check: if content is too large or contains many null bytes, treat as binary
	if len(content) > 50000 || isBinaryContent(content) {
		return "", fmt.Errorf("file too large or appears to be binary: %s", params.Path)
	}

	return string(content), nil
}

func (a *Agent) listFiles(input json.RawMessage) (string, error) {
	var params ListFilesInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %v", params)
	}

	entries, err := os.ReadDir(params.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %v", params.Path, err)
	}

	var result strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			result.WriteString(fmt.Sprintf("[DIR] %s\n", entry.Name()))
		} else {
			result.WriteString(fmt.Sprintf("[FILE] %s\n", entry.Name()))
		}
	}

	return result.String(), nil
}

func (a *Agent) editFile(input json.RawMessage) (string, error) {
	var params EditFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	content, err := os.ReadFile(params.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", params.Path, err)
	}

	oldContent := string(content)
	if !strings.Contains(oldContent, params.OldContent) {
		return "", fmt.Errorf("old content not found in file %s", params.Path)
	}

	newContent := strings.Replace(oldContent, params.OldContent, params.NewContent, 1)

	err = os.WriteFile(params.Path, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %v", params.Path, err)
	}

	return fmt.Sprintf("Successfully edited file %s", params.Path), nil
}

func (a *Agent) createFile(input json.RawMessage) (string, error) {
	var params CreateFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Check if file already exists
	if _, err := os.Stat(params.Path); err == nil {
		return "", fmt.Errorf("file already exists: %s", params.Path)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(params.Path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create and write the file
	err := os.WriteFile(params.Path, []byte(params.Content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", params.Path, err)
	}

	return fmt.Sprintf("Successfully created file %s", params.Path), nil
}

func isBinaryFile(path string) bool {
	// Check common binary file extensions
	binaryExts := []string{".exe", ".bin", ".so", ".dylib", ".dll", ".o", ".a", ".zip", ".tar", ".gz", ".pdf", ".jpg", ".png", ".gif", ".mp3", ".mp4", ".avi"}
	for _, ext := range binaryExts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}
	
	// Check if file has no extension (often executables on Unix)
	base := filepath.Base(path)
	return !strings.Contains(base, ".")
}

func isBinaryContent(content []byte) bool {
	// If more than 10% of first 1024 bytes are null or non-printable, likely binary
	checkSize := 1024
	if len(content) < checkSize {
		checkSize = len(content)
	}
	
	nullCount := 0
	for i := 0; i < checkSize; i++ {
		if content[i] == 0 || (content[i] < 32 && content[i] != 9 && content[i] != 10 && content[i] != 13) {
			nullCount++
		}
	}
	
	return nullCount > checkSize/10
}