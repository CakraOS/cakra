package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTransactionExecuteRollbackOnFailure(t *testing.T) {
	staging := t.TempDir()
	root := t.TempDir()
	backup := t.TempDir()

	// Existing file.
	existing := filepath.Join(
		root,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		existing,
		"old hello",
		0o755,
	)

	// File that should be created.
	newFile := filepath.Join(
		staging,
		"usr",
		"bin",
		"cakra",
	)

	writeTestFile(
		t,
		newFile,
		"cakra",
		0o755,
	)

	// Existing file that will be overwritten.
	overwrite := filepath.Join(
		staging,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		overwrite,
		"new hello",
		0o755,
	)

	// Force failure when committing "hello".
	commitHook = func(src, dst string) error {
		if filepath.Base(dst) == "hello" {
			return errors.New(
				"injected commit failure",
			)
		}

		return nil
	}

	defer func() {
		commitHook = nil
	}()

	tx := New(
		staging,
		root,
		backup,
	)

	err := tx.Execute()

	if err == nil {
		t.Fatal(
			"expected transaction to fail",
		)
	}

	// The original file must be restored.
	assertFileContent(
		t,
		existing,
		"old hello",
	)

	// The new file must not remain.
	if _, err := os.Stat(
		filepath.Join(
			root,
			"usr",
			"bin",
			"cakra",
		),
	); !os.IsNotExist(err) {
		t.Fatal(
			"new file remained after rollback",
		)
	}

	// Staging must be cleaned.
	if _, err := os.Stat(staging); !os.IsNotExist(err) {
		t.Fatal(
			"staging directory was not cleaned",
		)
	}

	// Backup must be cleaned.
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Fatal(
			"backup directory was not cleaned",
		)
	}
}
