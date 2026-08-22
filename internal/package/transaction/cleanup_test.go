package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupRemovesStagingAndBackup(t *testing.T) {
	staging := t.TempDir()
	backup := t.TempDir()
	root := t.TempDir()

	tx := New(
		staging,
		root,
		backup,
	)

	if err := tx.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}

	if _, err := os.Stat(staging); !os.IsNotExist(err) {
		t.Fatal(
			"staging still exists",
		)
	}

	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Fatal(
			"backup still exists",
		)
	}
}

func TestCleanupCanRunTwice(t *testing.T) {
	staging := filepath.Join(
		t.TempDir(),
		"staging",
	)

	backup := filepath.Join(
		t.TempDir(),
		"backup",
	)

	root := t.TempDir()

	if err := os.MkdirAll(
		staging,
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(
		backup,
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	tx := New(
		staging,
		root,
		backup,
	)

	if err := tx.Cleanup(); err != nil {
		t.Fatalf(
			"first cleanup failed: %v",
			err,
		)
	}

	if err := tx.Cleanup(); err != nil {
		t.Fatalf(
			"second cleanup failed: %v",
			err,
		)
	}
}

func TestCleanupRejectsFilesystemRoot(t *testing.T) {
	tx := New(
		string(os.PathSeparator),
		t.TempDir(),
		t.TempDir(),
	)

	if err := tx.Cleanup(); err == nil {
		t.Fatal(
			"cleanup should reject filesystem root",
		)
	}
}

func TestCleanupRejectsCurrentDirectory(t *testing.T) {
	tx := New(
		".",
		t.TempDir(),
		t.TempDir(),
	)

	if err := tx.Cleanup(); err == nil {
		t.Fatal(
			"cleanup should reject current directory",
		)
	}
}

func TestCleanupAllowsEmptyPaths(t *testing.T) {
	tx := New(
		"",
		t.TempDir(),
		"",
	)

	if err := tx.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup with empty paths failed: %v",
			err,
		)
	}
}
