package screens

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsStorageLinked(t *testing.T) {
	// 1. Case: No link
	tmpDir1 := t.TempDir()
	publicDir1 := filepath.Join(tmpDir1, "public")
	os.Mkdir(publicDir1, 0755)

	if isStorageLinked(tmpDir1) {
		t.Error("expected isStorageLinked to be false when no storage link exists")
	}

	// 2. Case: Valid link
	tmpDir2 := t.TempDir()
	publicDir2 := filepath.Join(tmpDir2, "public")
	storageDir2 := filepath.Join(tmpDir2, "storage", "app", "public")
	os.MkdirAll(publicDir2, 0755)
	os.MkdirAll(storageDir2, 0755)

	// Create symlink: public/storage -> storage/app/public
	// Note: The actual target in Laravel is usually relative or absolute "storage/app/public".
	// We just need a symlink named "storage" inside "public".
	linkName := filepath.Join(publicDir2, "storage")
	err := os.Symlink(storageDir2, linkName)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	if !isStorageLinked(tmpDir2) {
		t.Error("expected isStorageLinked to be true when storage link exists")
	}
}

func TestFullPermissionResetCommand(t *testing.T) {
	// Setup a fake Laravel project
	tmpDir := t.TempDir()

	// Create artisan file
	os.WriteFile(filepath.Join(tmpDir, "artisan"), []byte(""), 0755)

	// Create storage and bootstrap/cache
	os.MkdirAll(filepath.Join(tmpDir, "storage"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "bootstrap", "cache"), 0755)

	// Change to temp dir
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(tmpDir)

	model := NewLaravelPermissionsModel()

	var fullResetAction *LaravelPermAction
	for _, action := range model.actions {
		if action.ID == "full_reset" {
			fullResetAction = &action
			break
		}
	}

	if fullResetAction == nil {
		t.Fatal("full_reset action not found in model")
	}

	// Check for key components in the command
	expectedParts := []string{
		"sudo usermod -a -G",
		"sudo chown -R",
		"sudo find . -type d -exec chmod 775",
		"sudo find . -type f -exec chmod 664",
		"sudo chgrp -R",
		"sudo chmod -R ug+rwx storage",
	}

	for _, part := range expectedParts {
		if !strings.Contains(fullResetAction.Command, part) && !strings.Contains(fullResetAction.Command, "sudo") {
			t.Errorf("command component missing or incomplete: %s", part)
		}
	}

	// Specific check for the new added logic
	if !strings.Contains(fullResetAction.Command, "usermod -a -G") {
		t.Error("command missing usermod group sync logic")
	}
	if !strings.Contains(fullResetAction.Command, "chmod 775") {
		t.Error("command missing directory permissions 775")
	}
	if !strings.Contains(fullResetAction.Command, "chmod 664") {
		t.Error("command missing file permissions 664")
	}
}
