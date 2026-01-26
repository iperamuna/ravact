package executor

import (
	"embed"
	"io"
	"testing"
	"testing/fstest"
	"time"
)

// Create a test embedded filesystem
func createTestFS() embed.FS {
	// Note: We can't easily create an embed.FS at runtime for testing
	// So we'll test what we can without actual embedding
	return embed.FS{}
}

func TestNewScriptRunner(t *testing.T) {
	fs := embed.FS{}
	runner := NewScriptRunner(fs)

	if runner == nil {
		t.Fatal("NewScriptRunner returned nil")
	}
}

func TestBytesReader(t *testing.T) {
	testData := []byte("Hello, World!")
	reader := bytesReader(testData)

	if reader == nil {
		t.Fatal("bytesReader returned nil")
	}

	// Read all data
	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if string(result) != string(testData) {
		t.Errorf("expected '%s', got '%s'", string(testData), string(result))
	}
}

func TestBytesReaderWrapper_Read(t *testing.T) {
	testData := []byte("Test data for reading")
	wrapper := &bytesReaderWrapper{data: testData, pos: 0}

	// Read in chunks
	buf := make([]byte, 5)

	// First read
	n, err := wrapper.Read(buf)
	if err != nil {
		t.Fatalf("first read failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected to read 5 bytes, got %d", n)
	}
	if string(buf[:n]) != "Test " {
		t.Errorf("expected 'Test ', got '%s'", string(buf[:n]))
	}

	// Continue reading
	n, err = wrapper.Read(buf)
	if err != nil {
		t.Fatalf("second read failed: %v", err)
	}
	if string(buf[:n]) != "data " {
		t.Errorf("expected 'data ', got '%s'", string(buf[:n]))
	}
}

func TestBytesReaderWrapper_EOF(t *testing.T) {
	testData := []byte("Hi")
	wrapper := &bytesReaderWrapper{data: testData, pos: 0}

	buf := make([]byte, 10)

	// First read - should get all data
	n, err := wrapper.Read(buf)
	if err != nil {
		t.Fatalf("first read failed: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 bytes, got %d", n)
	}

	// Second read - should get EOF
	n, err = wrapper.Read(buf)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes at EOF, got %d", n)
	}
}

func TestBytesReaderWrapper_EmptyData(t *testing.T) {
	wrapper := &bytesReaderWrapper{data: []byte{}, pos: 0}

	buf := make([]byte, 10)
	n, err := wrapper.Read(buf)

	if err != io.EOF {
		t.Errorf("expected EOF for empty data, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes for empty data, got %d", n)
	}
}

func TestBytesReaderWrapper_LargeData(t *testing.T) {
	// Create large test data
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	wrapper := &bytesReaderWrapper{data: largeData, pos: 0}

	// Read all data
	result, err := io.ReadAll(wrapper)
	if err != nil {
		t.Fatalf("failed to read large data: %v", err)
	}

	if len(result) != len(largeData) {
		t.Errorf("expected %d bytes, got %d", len(largeData), len(result))
	}

	// Verify data integrity
	for i := range result {
		if result[i] != largeData[i] {
			t.Errorf("data mismatch at position %d", i)
			break
		}
	}
}

// TestScriptRunnerWithMockFS tests the runner with a mock filesystem
// Note: This is limited because embed.FS can't be created at runtime
func TestScriptRunnerWithMockFS(t *testing.T) {
	// We can use fstest.MapFS for testing patterns
	mockFS := fstest.MapFS{
		"scripts/test.sh": &fstest.MapFile{
			Data: []byte("#!/bin/bash\necho 'Hello'"),
		},
	}

	// Verify our mock works
	data, err := mockFS.ReadFile("scripts/test.sh")
	if err != nil {
		t.Fatalf("mock FS read failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty script content")
	}
}

func TestScriptRunner_ExecuteScript_NotFound(t *testing.T) {
	// With an empty embed.FS, any script should fail to read
	runner := NewScriptRunner(embed.FS{})

	_, err := runner.ExecuteScript("nonexistent.sh", 5*time.Second)
	if err == nil {
		t.Error("expected error for non-existent script")
	}
}

func TestScriptRunner_TimeoutDuration(t *testing.T) {
	// Test that timeout parameter is accepted
	runner := NewScriptRunner(embed.FS{})

	// These should not panic even though they'll fail (no scripts in empty FS)
	timeouts := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		1 * time.Minute,
	}

	for _, timeout := range timeouts {
		_, err := runner.ExecuteScript("test.sh", timeout)
		if err == nil {
			t.Errorf("expected error with timeout %v (empty FS)", timeout)
		}
	}
}
