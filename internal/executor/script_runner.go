package executor

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// ScriptRunner executes scripts from embedded filesystem
type ScriptRunner struct {
	embeddedFS embed.FS
}

// NewScriptRunner creates a new script runner
func NewScriptRunner(fs embed.FS) *ScriptRunner {
	return &ScriptRunner{embeddedFS: fs}
}

// ExecuteScript runs a script from embedded FS by piping it to bash
func (sr *ScriptRunner) ExecuteScript(scriptPath string, timeout time.Duration) (string, error) {
	// Read script content from embedded FS
	scriptContent, err := sr.embeddedFS.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded script %s: %w", scriptPath, err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute bash with script content piped to stdin
	cmd := exec.CommandContext(ctx, "bash", "-s")
	cmd.Stdin = bytesReader(scriptContent)

	// Capture combined output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("script execution failed: %w", err)
	}

	return string(output), nil
}

// bytesReader creates an io.Reader from byte slice
func bytesReader(data []byte) io.Reader {
	return &bytesReaderWrapper{data: data, pos: 0}
}

type bytesReaderWrapper struct {
	data []byte
	pos  int
}

func (r *bytesReaderWrapper) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
