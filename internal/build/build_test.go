package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStagingPath(t *testing.T) {
	root := t.TempDir()

	buildDir := filepath.Join(root, "build", "hello")
	destDir := filepath.Join(buildDir, "dest")

	if err := os.MkdirAll(
		filepath.Join(destDir, "usr", "bin"),
		0755,
	); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(
		destDir,
		"usr",
		"bin",
		"hello",
	)

	if err := os.WriteFile(
		testFile,
		[]byte("test"),
		0755,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("staged file does not exist: %v", err)
	}
}
