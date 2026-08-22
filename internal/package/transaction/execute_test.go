package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTransactionExecute(t *testing.T) {
	staging := t.TempDir()
	root := t.TempDir()
	backup := t.TempDir()

	src := filepath.Join(
		staging,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		src,
		"hello from cakra",
		0o755,
	)

	tx := New(
		staging,
		root,
		backup,
	)

	if err := tx.Execute(); err != nil {
		t.Fatalf(
			"transaction failed: %v",
			err,
		)
	}

	dst := filepath.Join(
		root,
		"usr",
		"bin",
		"hello",
	)

	assertFileContent(
		t,
		dst,
		"hello from cakra",
	)

	if _, err := os.Stat(staging); !os.IsNotExist(err) {
		t.Fatalf(
			"staging directory was not cleaned",
		)
	}

	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Fatalf(
			"backup directory was not cleaned",
		)
	}
}
