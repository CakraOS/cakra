package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommitAndRollbackNewFile(t *testing.T) {
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
		"hello world",
		0755,
	)

	journal, err := Commit(
		staging,
		root,
		backup,
	)
	if err != nil {
		t.Fatalf(
			"commit failed: %v",
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
		"hello world",
	)

	if err := Rollback(journal); err != nil {
		t.Fatalf(
			"rollback failed: %v",
			err,
		)
	}

	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf(
			"file still exists after rollback",
		)
	}
}

func TestCommitAndRollbackOverwrite(t *testing.T) {
	staging := t.TempDir()
	root := t.TempDir()
	backup := t.TempDir()

	dst := filepath.Join(
		root,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		dst,
		"old version",
		0755,
	)

	src := filepath.Join(
		staging,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		src,
		"new version",
		0755,
	)

	journal, err := Commit(
		staging,
		root,
		backup,
	)
	if err != nil {
		t.Fatalf(
			"commit failed: %v",
			err,
		)
	}

	assertFileContent(
		t,
		dst,
		"new version",
	)

	if err := Rollback(journal); err != nil {
		t.Fatalf(
			"rollback failed: %v",
			err,
		)
	}

	assertFileContent(
		t,
		dst,
		"old version",
	)
}

func TestCommitAndRollbackMultipleFiles(t *testing.T) {
	staging := t.TempDir()
	root := t.TempDir()
	backup := t.TempDir()

	files := map[string]string{
		"usr/bin/hello":   "hello",
		"usr/bin/cakra":   "cakra",
		"usr/share/info": "information",
	}

	for path, content := range files {
		writeTestFile(
			t,
			filepath.Join(
				staging,
				path,
			),
			content,
			0644,
		)
	}

	journal, err := Commit(
		staging,
		root,
		backup,
	)
	if err != nil {
		t.Fatalf(
			"commit failed: %v",
			err,
		)
	}

	for path, content := range files {
		assertFileContent(
			t,
			filepath.Join(root, path),
			content,
		)
	}

	if err := Rollback(journal); err != nil {
		t.Fatalf(
			"rollback failed: %v",
			err,
		)
	}

	for path := range files {
		fullPath := filepath.Join(
			root,
			path,
		)

		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			t.Fatalf(
				"file still exists after rollback: %s",
				fullPath,
			)
		}
	}
}

func TestCommitAndRollbackDirectories(t *testing.T) {
	staging := t.TempDir()
	root := t.TempDir()
	backup := t.TempDir()

	src := filepath.Join(
		staging,
		"usr",
		"share",
		"cakra",
		"data.txt",
	)

	writeTestFile(
		t,
		src,
		"cakra data",
		0644,
	)

	journal, err := Commit(
		staging,
		root,
		backup,
	)
	if err != nil {
		t.Fatalf(
			"commit failed: %v",
			err,
		)
	}

	dir := filepath.Join(
		root,
		"usr",
		"share",
		"cakra",
	)

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf(
			"directory was not created: %v",
			err,
		)
	}

	if err := Rollback(journal); err != nil {
		t.Fatalf(
			"rollback failed: %v",
			err,
		)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf(
			"directory still exists after rollback",
		)
	}
}

func writeTestFile(
	t *testing.T,
	path string,
	content string,
	mode os.FileMode,
) {
	t.Helper()

	if err := os.MkdirAll(
		filepath.Dir(path),
		0755,
	); err != nil {
		t.Fatalf(
			"create test directory: %v",
			err,
		)
	}

	if err := os.WriteFile(
		path,
		[]byte(content),
		mode,
	); err != nil {
		t.Fatalf(
			"write test file: %v",
			err,
		)
	}
}

func assertFileContent(
	t *testing.T,
	path string,
	expected string,
) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			path,
			err,
		)
	}

	if string(data) != expected {
		t.Fatalf(
			"unexpected content in %s: got %q, want %q",
			path,
			string(data),
			expected,
		)
	}
}
