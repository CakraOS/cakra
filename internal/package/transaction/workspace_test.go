package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewWorkspace(t *testing.T) {
	parent := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	if workspace.Root == "" {
		t.Fatal("workspace root is empty")
	}

	if workspace.Staging == "" {
		t.Fatal("workspace staging is empty")
	}

	if workspace.Backup == "" {
		t.Fatal("workspace backup is empty")
	}

	if _, err := os.Stat(
		workspace.Root,
	); err != nil {
		t.Fatalf(
			"workspace root does not exist: %v",
			err,
		)
	}

	if _, err := os.Stat(
		workspace.Staging,
	); err != nil {
		t.Fatalf(
			"staging does not exist: %v",
			err,
		)
	}

	if _, err := os.Stat(
		workspace.Backup,
	); err != nil {
		t.Fatalf(
			"backup does not exist: %v",
			err,
		)
	}

	if filepath.Base(
		workspace.Root,
	)[:len("cakra-txn-")] != "cakra-txn-" {
		t.Fatalf(
			"invalid transaction name: %s",
			workspace.Root,
		)
	}
}

func TestWorkspaceCleanup(t *testing.T) {
	parent := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	writeTestFile(
		t,
		filepath.Join(
			workspace.Staging,
			"test.txt",
		),
		"staging",
		0o644,
	)

	writeTestFile(
		t,
		filepath.Join(
			workspace.Backup,
			"test.txt",
		),
		"backup",
		0o644,
	)

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"workspace cleanup failed: %v",
			err,
		)
	}

	if _, err := os.Stat(
		workspace.Root,
	); !os.IsNotExist(err) {
		t.Fatal(
			"workspace still exists after cleanup",
		)
	}
}

func TestWorkspaceCleanupTwice(t *testing.T) {
	parent := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"first cleanup failed: %v",
			err,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"second cleanup failed: %v",
			err,
		)
	}
}

func TestWorkspaceRejectsFilesystemRoot(t *testing.T) {
	_, err := NewWorkspace("/")

	if err == nil {
		t.Fatal(
			"filesystem root should be rejected",
		)
	}
}

func TestNewFromWorkspace(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	tx := NewFromWorkspace(
		workspace,
		root,
	)

	if tx.Staging != workspace.Staging {
		t.Fatal(
			"transaction staging does not match workspace",
		)
	}

	if tx.Backup != workspace.Backup {
		t.Fatal(
			"transaction backup does not match workspace",
		)
	}

	if tx.Root != root {
		t.Fatal(
			"transaction root does not match",
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}

func TestWorkspaceTransaction(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	src := filepath.Join(
		workspace.Staging,
		"usr",
		"bin",
		"cakra",
	)

	writeTestFile(
		t,
		src,
		"hello from cakra",
		0o755,
	)

	tx := NewFromWorkspace(
		workspace,
		root,
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
		"cakra",
	)

	assertFileContent(
		t,
		dst,
		"hello from cakra",
	)

	if _, err := os.Stat(
		workspace.Root,
	); !os.IsNotExist(err) {
		t.Fatal(
			"workspace was not cleaned",
		)
	}
}

func TestNewWorkspaceCreatesPersistentState(t *testing.T) {
	parent := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	state, err := readState(workspace.Root)
	if err != nil {
		t.Fatalf(
			"read workspace state: %v",
			err,
		)
	}

	if state != StateCreated {
		t.Fatalf(
			"expected state %q, got %q",
			StateCreated,
			state,
		)
	}

	if workspace.State != StateCreated {
		t.Fatalf(
			"expected workspace state %q, got %q",
			StateCreated,
			workspace.State,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}
