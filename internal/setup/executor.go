package setup

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/iperamuna/ravact/internal/models"
)

// Executor handles the execution of setup scripts
type Executor struct {
	scriptsDir string
	logWriter  io.Writer
}

// NewExecutor creates a new setup script executor
func NewExecutor(scriptsDir string) *Executor {
	return &Executor{
		scriptsDir: scriptsDir,
		logWriter:  os.Stdout,
	}
}

// SetLogWriter sets the log output writer
func (e *Executor) SetLogWriter(w io.Writer) {
	e.logWriter = w
}

// ExecuteScript executes a setup script
func (e *Executor) ExecuteScript(script models.SetupScript) (*models.ExecutionResult, error) {
	startTime := time.Now()
	result := &models.ExecutionResult{
		Timestamp: startTime,
	}

	// Validate script exists
	scriptPath := script.ScriptPath
	if !strings.HasPrefix(scriptPath, "/") {
		scriptPath = fmt.Sprintf("%s/%s", e.scriptsDir, scriptPath)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		result.Success = false
		result.Error = fmt.Sprintf("Script not found: %s", scriptPath)
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("script not found: %s", scriptPath)
	}

	// Check if script is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		return result, err
	}

	if info.Mode()&0111 == 0 {
		// Make script executable
		if err := os.Chmod(scriptPath, 0755); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Cannot make script executable: %v", err)
			result.Duration = time.Since(startTime)
			return result, err
		}
	}

	// Set timeout
	timeout := script.Timeout
	if timeout == 0 {
		timeout = 30 * time.Minute // Default 30 minutes
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range script.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		return result, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Start command
	if err := cmd.Start(); err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		return result, err
	}

	// Read output
	outputChan := make(chan string)
	go e.readOutput(stdout, outputChan)
	go e.readOutput(stderr, outputChan)

	var outputLines []string
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// Collect output
	for {
		select {
		case line := <-outputChan:
			outputLines = append(outputLines, line)
			if e.logWriter != nil {
				fmt.Fprintln(e.logWriter, line)
			}
		case err := <-done:
			// Drain remaining output
			for {
				select {
				case line := <-outputChan:
					outputLines = append(outputLines, line)
					if e.logWriter != nil {
						fmt.Fprintln(e.logWriter, line)
					}
				default:
					goto finished
				}
			}
		finished:
			result.Duration = time.Since(startTime)
			result.Output = strings.Join(outputLines, "\n")

			if err != nil {
				result.Success = false
				if exitErr, ok := err.(*exec.ExitError); ok {
					result.ExitCode = exitErr.ExitCode()
				}
				result.Error = err.Error()
				return result, err
			}

			result.Success = true
			result.ExitCode = 0
			return result, nil
		case <-ctx.Done():
			result.Duration = time.Since(startTime)
			result.Success = false
			result.Error = "Script execution timed out"
			result.Output = strings.Join(outputLines, "\n")
			return result, fmt.Errorf("script execution timed out")
		}
	}
}

// readOutput reads output line by line
func (e *Executor) readOutput(reader io.Reader, output chan<- string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		output <- scanner.Text()
	}
}

// ValidateScript validates a script before execution
func (e *Executor) ValidateScript(script models.SetupScript) error {
	scriptPath := script.ScriptPath
	if !strings.HasPrefix(scriptPath, "/") {
		scriptPath = fmt.Sprintf("%s/%s", e.scriptsDir, scriptPath)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// Check if it's a bash script
	file, err := os.Open(scriptPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := scanner.Text()
		if !strings.HasPrefix(firstLine, "#!/bin/bash") && !strings.HasPrefix(firstLine, "#!/bin/sh") {
			return fmt.Errorf("script must start with #!/bin/bash or #!/bin/sh")
		}
	}

	return nil
}

// GetAvailableScripts returns all available setup scripts
func (e *Executor) GetAvailableScripts() ([]models.SetupScript, error) {
	entries, err := os.ReadDir(e.scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Scripts directory doesn't exist yet
			return []models.SetupScript{}, nil
		}
		return nil, err
	}

	var scripts []models.SetupScript
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sh") {
			continue
		}

		scripts = append(scripts, models.SetupScript{
			ID:         strings.TrimSuffix(entry.Name(), ".sh"),
			Name:       strings.TrimSuffix(entry.Name(), ".sh"),
			ScriptPath: entry.Name(),
		})
	}

	return scripts, nil
}
