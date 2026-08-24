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

func TestTransactionMarkStaged(t *testing.T) {
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

	if workspace.State != StateCreated {
		t.Fatalf(
			"expected initial state %q, got %q",
			StateCreated,
			workspace.State,
		)
	}

	if err := tx.MarkStaged(); err != nil {
		t.Fatalf(
			"mark staged failed: %v",
			err,
		)
	}

	if workspace.State != StateStaged {
		t.Fatalf(
			"expected state %q, got %q",
			StateStaged,
			workspace.State,
		)
	}

	state, err := readState(workspace.Root)
	if err != nil {
		t.Fatalf(
			"read state failed: %v",
			err,
		)
	}

	if state != StateStaged {
		t.Fatalf(
			"expected persistent state %q, got %q",
			StateStaged,
			state,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}

func TestTransactionMarkStagedRejectsInvalidState(t *testing.T) {
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

	workspace.State = StateStaged

	if err := tx.MarkStaged(); err == nil {
		t.Fatal(
			"expected invalid state transition",
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}

func TestTransactionCommitStateTransition(t *testing.T) {
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

	if err := tx.MarkCommitting(); err == nil {
		t.Fatal(
			"expected committing transition to reject CREATED state",
		)
	}

	if err := tx.MarkStaged(); err != nil {
		t.Fatalf(
			"mark staged failed: %v",
			err,
		)
	}

	if err := tx.MarkCommitting(); err != nil {
		t.Fatalf(
			"mark committing failed: %v",
			err,
		)
	}

	if workspace.State != StateCommitting {
		t.Fatalf(
			"expected state %q, got %q",
			StateCommitting,
			workspace.State,
		)
	}

	state, err := readState(workspace.Root)
	if err != nil {
		t.Fatalf(
			"read state failed: %v",
			err,
		)
	}

	if state != StateCommitting {
		t.Fatalf(
			"expected persistent state %q, got %q",
			StateCommitting,
			state,
		)
	}

	if err := tx.MarkCommitted(); err != nil {
		t.Fatalf(
			"mark committed failed: %v",
			err,
		)
	}

	if workspace.State != StateCommitted {
		t.Fatalf(
			"expected state %q, got %q",
			StateCommitted,
			workspace.State,
		)
	}

	state, err = readState(workspace.Root)
	if err != nil {
		t.Fatalf(
			"read committed state failed: %v",
			err,
		)
	}

	if state != StateCommitted {
		t.Fatalf(
			"expected persistent state %q, got %q",
			StateCommitted,
			state,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}

func TestTransactionFailureRollbackStateTransition(t *testing.T) {
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

	if err := tx.MarkFailed(); err == nil {
		t.Fatal(
			"expected FAILED transition to reject CREATED state",
		)
	}

	if err := tx.MarkStaged(); err != nil {
		t.Fatalf(
			"mark staged failed: %v",
			err,
		)
	}

	if err := tx.MarkCommitting(); err != nil {
		t.Fatalf(
			"mark committing failed: %v",
			err,
		)
	}

	if err := tx.MarkFailed(); err != nil {
		t.Fatalf(
			"mark failed failed: %v",
			err,
		)
	}

	if workspace.State != StateFailed {
		t.Fatalf(
			"expected state %q, got %q",
			StateFailed,
			workspace.State,
		)
	}

	if err := tx.MarkRollingBack(); err != nil {
		t.Fatalf(
			"mark rolling back failed: %v",
			err,
		)
	}

	if workspace.State != StateRollingBack {
		t.Fatalf(
			"expected state %q, got %q",
			StateRollingBack,
			workspace.State,
		)
	}

	if err := tx.MarkRolledBack(); err != nil {
		t.Fatalf(
			"mark rolled back failed: %v",
			err,
		)
	}

	if workspace.State != StateRolledBack {
		t.Fatalf(
			"expected state %q, got %q",
			StateRolledBack,
			workspace.State,
		)
	}

	state, err := readState(workspace.Root)
	if err != nil {
		t.Fatalf(
			"read final state failed: %v",
			err,
		)
	}

	if state != StateRolledBack {
		t.Fatalf(
			"expected persistent state %q, got %q",
			StateRolledBack,
			state,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}

func TestTransactionCommitFailureMarksFailed(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

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
			"usr",
			"bin",
			"test",
		),
		"test",
		0o755,
	)

	tx := NewFromWorkspace(
		workspace,
		root,
	)

	if err := tx.MarkStaged(); err != nil {
		t.Fatalf(
			"mark staged failed: %v",
			err,
		)
	}

	commitHook = func(src, dst string) error {
		return errors.New(
			"injected commit failure",
		)
	}

	defer func() {
		commitHook = nil
	}()

	if err := tx.Commit(); err == nil {
		t.Fatal(
			"expected commit failure",
		)
	}

	if workspace.State != StateFailed {
		t.Fatalf(
			"expected state %q, got %q",
			StateFailed,
			workspace.State,
		)
	}

	state, err := readState(workspace.Root)
	if err != nil {
		t.Fatalf(
			"read state failed: %v",
			err,
		)
	}

	if state != StateFailed {
		t.Fatalf(
			"expected persistent state %q, got %q",
			StateFailed,
			state,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}

func TestTransactionRollbackStateTransition(t *testing.T) {
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

	if err := tx.MarkStaged(); err != nil {
		t.Fatalf(
			"mark staged failed: %v",
			err,
		)
	}

	if err := tx.MarkCommitting(); err != nil {
		t.Fatalf(
			"mark committing failed: %v",
			err,
		)
	}

	if err := tx.MarkFailed(); err != nil {
		t.Fatalf(
			"mark failed failed: %v",
			err,
		)
	}

	tx.Journal = &Journal{}

	if err := tx.Rollback(); err != nil {
		t.Fatalf(
			"rollback failed: %v",
			err,
		)
	}

	if workspace.State != StateRolledBack {
		t.Fatalf(
			"expected state %q, got %q",
			StateRolledBack,
			workspace.State,
		)
	}

	state, err := readState(workspace.Root)
	if err != nil {
		t.Fatalf(
			"read state failed: %v",
			err,
		)
	}

	if state != StateRolledBack {
		t.Fatalf(
			"expected persistent state %q, got %q",
			StateRolledBack,
			state,
		)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup failed: %v",
			err,
		)
	}
}
