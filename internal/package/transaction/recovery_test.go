package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindIncompleteTransactions(t *testing.T) {
	parent := t.TempDir()

	created, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace: %v",
			err,
		)
	}

	committed, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create committed workspace: %v",
			err,
		)
	}

	if err := writeState(
		committed.Root,
		StateCommitted,
	); err != nil {
		t.Fatalf(
			"write committed state: %v",
			err,
		)
	}

	incomplete, err := FindIncompleteTransactions(parent)
	if err != nil {
		t.Fatalf(
			"find incomplete transactions: %v",
			err,
		)
	}

	if len(incomplete) != 1 {
		t.Fatalf(
			"expected 1 incomplete transaction, got %d",
			len(incomplete),
		)
	}

	if incomplete[0] != created.Root {
		t.Fatalf(
			"unexpected incomplete transaction: %s",
			incomplete[0],
		)
	}
}

func TestRecoverCreatedWorkspace(t *testing.T) {
	parent := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace: %v",
			err,
		)
	}

	if err := RecoverWorkspace(
		workspace.Root,
	); err != nil {
		t.Fatalf(
			"recover created workspace: %v",
			err,
		)
	}

	if _, err := os.Stat(
		workspace.Root,
	); !os.IsNotExist(err) {
		t.Fatal(
			"workspace still exists after recovery",
		)
	}
}

func TestRecoverCommittedWorkspace(t *testing.T) {
	parent := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace: %v",
			err,
		)
	}

	if err := writeState(
		workspace.Root,
		StateCommitted,
	); err != nil {
		t.Fatalf(
			"write committed state: %v",
			err,
		)
	}

	if err := RecoverWorkspace(
		workspace.Root,
	); err != nil {
		t.Fatalf(
			"recover committed workspace: %v",
			err,
		)
	}

	if _, err := os.Stat(
		workspace.Root,
	); !os.IsNotExist(err) {
		t.Fatal(
			"committed workspace still exists",
		)
	}
}

func TestRecoverCommittingWorkspace(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace: %v",
			err,
		)
	}

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

	newFile := filepath.Join(
		workspace.Staging,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		newFile,
		"new hello",
		0o755,
	)

	journal, err := Commit(
		workspace.Staging,
		root,
		workspace.Backup,
	)
	if err != nil {
		t.Fatalf(
			"commit transaction: %v",
			err,
		)
	}

	if journal == nil {
		t.Fatal(
			"journal is nil",
		)
	}

	if err := writeJournal(
		journalPath(workspace.Root),
		journal,
	); err != nil {
		t.Fatalf(
			"write journal: %v",
			err,
		)
	}

	if err := writeState(
		workspace.Root,
		StateCommitting,
	); err != nil {
		t.Fatalf(
			"write committing state: %v",
			err,
		)
	}

	assertFileContent(
		t,
		existing,
		"new hello",
	)

	if err := RecoverWorkspace(
		workspace.Root,
	); err != nil {
		t.Fatalf(
			"recover committing workspace: %v",
			err,
		)
	}

	assertFileContent(
		t,
		existing,
		"old hello",
	)

	if _, err := os.Stat(
		workspace.Root,
	); !os.IsNotExist(err) {
		t.Fatal(
			"workspace still exists after recovery",
		)
	}
}

func TestRecoverRejectsUnsafePath(t *testing.T) {
	err := RecoverWorkspace(".")

	if err == nil {
		t.Fatal(
			"expected unsafe path to be rejected",
		)
	}
}

func TestFindIncompleteTransactionsMissingParent(
	t *testing.T,
) {
	parent := filepath.Join(
		t.TempDir(),
		"does-not-exist",
	)

	transactions, err := FindIncompleteTransactions(
		parent,
	)
	if err != nil {
		t.Fatalf(
			"find transactions failed: %v",
			err,
		)
	}

	if len(transactions) != 0 {
		t.Fatalf(
			"expected no transactions, got %d",
			len(transactions),
		)
	}
}

func TestRecoverFailedWorkspace(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf("create workspace failed: %v", err)
	}

	original := filepath.Join(root, "usr", "bin", "hello")
	backup := filepath.Join(
		workspace.Backup,
		backupName(original),
	)

	writeTestFile(t, original, "new hello", 0o755)
	writeTestFile(t, backup, "old hello", 0o755)

	journal := &Journal{
		Backups: []Backup{
			{
				Original: original,
				Backup:   backup,
			},
		},
	}

	if err := writeJournal(
		journalPath(workspace.Root),
		journal,
	); err != nil {
		t.Fatalf("write journal failed: %v", err)
	}

	if err := writeState(
		workspace.Root,
		StateFailed,
	); err != nil {
		t.Fatalf("write failed state: %v", err)
	}

	workspace.State = StateFailed

	if err := RecoverWorkspace(workspace.Root); err != nil {
		t.Fatalf("recover failed workspace: %v", err)
	}

	assertFileContent(
		t,
		original,
		"old hello",
	)

	if _, err := os.Stat(workspace.Root); !os.IsNotExist(err) {
		t.Fatal("recovered workspace was not cleaned")
	}
}

func TestRecoverRollingBackWorkspace(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf("create workspace failed: %v", err)
	}

	original := filepath.Join(root, "usr", "bin", "hello")
	backup := filepath.Join(
		workspace.Backup,
		backupName(original),
	)

	writeTestFile(t, original, "new hello", 0o755)
	writeTestFile(t, backup, "old hello", 0o755)

	journal := &Journal{
		Backups: []Backup{
			{
				Original: original,
				Backup:   backup,
			},
		},
	}

	if err := writeJournal(
		journalPath(workspace.Root),
		journal,
	); err != nil {
		t.Fatalf("write journal failed: %v", err)
	}

	if err := writeState(
		workspace.Root,
		StateRollingBack,
	); err != nil {
		t.Fatalf("write rolling back state: %v", err)
	}

	workspace.State = StateRollingBack

	if err := RecoverWorkspace(workspace.Root); err != nil {
		t.Fatalf("recover rolling back workspace: %v", err)
	}

	assertFileContent(
		t,
		original,
		"old hello",
	)

	if _, err := os.Stat(workspace.Root); !os.IsNotExist(err) {
		t.Fatal("recovered workspace was not cleaned")
	}
}

func TestRecoverInterruptedRollback(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf("create workspace failed: %v", err)
	}

	originalA := filepath.Join(root, "usr", "bin", "a")
	originalB := filepath.Join(root, "usr", "bin", "b")

	backupA := filepath.Join(
		workspace.Backup,
		backupName(originalA),
	)
	backupB := filepath.Join(
		workspace.Backup,
		backupName(originalB),
	)

	writeTestFile(t, originalA, "new-a", 0o755)
	writeTestFile(t, originalB, "new-b", 0o755)

	writeTestFile(t, backupA, "old-a", 0o755)
	writeTestFile(t, backupB, "old-b", 0o755)

	journal := &Journal{
		Backups: []Backup{
			{
				Original: originalA,
				Backup:   backupA,
			},
			{
				Original: originalB,
				Backup:   backupB,
			},
		},
	}

	if err := writeJournal(
		journalPath(workspace.Root),
		journal,
	); err != nil {
		t.Fatalf("write journal failed: %v", err)
	}

	if err := writeState(
		workspace.Root,
		StateRollingBack,
	); err != nil {
		t.Fatalf("write rolling back state: %v", err)
	}

	/*
		Simulate a rollback that was already partially completed.

		File A has already been restored.
		File B is still in its committed state.
	*/
	if err := os.Remove(originalA); err != nil {
		t.Fatalf("remove original A failed: %v", err)
	}

	if err := copyRaw(backupA, originalA); err != nil {
		t.Fatalf("restore original A failed: %v", err)
	}

	// Recovery must safely repeat the rollback.
	if err := RecoverWorkspace(workspace.Root); err != nil {
		t.Fatalf(
			"first recovery failed: %v",
			err,
		)
	}

	assertFileContent(
		t,
		originalA,
		"old-a",
	)

	assertFileContent(
		t,
		originalB,
		"old-b",
	)

	if _, err := os.Stat(workspace.Root); !os.IsNotExist(err) {
		t.Fatal("workspace was not cleaned after recovery")
	}
}

func TestRecoverRolledBackWorkspaceIsIdempotent(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf("create workspace failed: %v", err)
	}

	original := filepath.Join(
		root,
		"usr",
		"bin",
		"hello",
	)

	backup := filepath.Join(
		workspace.Backup,
		backupName(original),
	)

	writeTestFile(t, original, "new hello", 0o755)
	writeTestFile(t, backup, "old hello", 0o755)

	journal := &Journal{
		Backups: []Backup{
			{
				Original: original,
				Backup:   backup,
			},
		},
	}

	if err := writeJournal(
		journalPath(workspace.Root),
		journal,
	); err != nil {
		t.Fatalf("write journal failed: %v", err)
	}

	if err := writeState(
		workspace.Root,
		StateRollingBack,
	); err != nil {
		t.Fatalf("write rolling back state: %v", err)
	}

	if err := RecoverWorkspace(workspace.Root); err != nil {
		t.Fatalf("first recovery failed: %v", err)
	}

	assertFileContent(
		t,
		original,
		"old hello",
	)

	/*
		The workspace has already been cleaned.
		A second recovery attempt must not turn
		a successfully recovered transaction into
		an error.
	*/
	if err := RecoverWorkspace(workspace.Root); err == nil {
		t.Fatal(
			"expected second recovery to report missing workspace",
		)
	}
}

func TestRecoveryAfterCrashDuringCommit(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	original := filepath.Join(
		root,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		original,
		"original",
		0o755,
	)

	replacement := filepath.Join(
		workspace.Staging,
		"usr",
		"bin",
		"hello",
	)

	writeTestFile(
		t,
		replacement,
		"replacement",
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

	if err := tx.MarkCommitting(); err != nil {
		t.Fatalf(
			"mark committing failed: %v",
			err,
		)
	}

	journal, err := Commit(
		tx.Staging,
		tx.Root,
		tx.Backup,
	)
	if err != nil {
		t.Fatalf(
			"commit failed: %v",
			err,
		)
	}

	if err := writeJournal(
		journalPath(workspace.Root),
		journal,
	); err != nil {
		t.Fatalf(
			"write journal failed: %v",
			err,
		)
	}

	// Simulate crash: the process disappears here.
	// The workspace remains in COMMITTING state.

	assertFileContent(
		t,
		original,
		"replacement",
	)

	manager := NewRecoveryManager(parent)

	if err := manager.Recover(); err != nil {
		t.Fatalf(
			"recovery failed: %v",
			err,
		)
	}

	// Recovery must restore the pre-transaction content.
	assertFileContent(
		t,
		original,
		"original",
	)

	if _, err := os.Stat(
		workspace.Root,
	); !os.IsNotExist(err) {
		t.Fatal(
			"transaction workspace remained after recovery",
		)
	}
}

func TestRecoveryAfterCrashWithCreatedFile(t *testing.T) {
	parent := t.TempDir()
	root := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	created := filepath.Join(
		workspace.Staging,
		"usr",
		"bin",
		"cakra",
	)

	writeTestFile(
		t,
		created,
		"cakra",
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

	if err := tx.MarkCommitting(); err != nil {
		t.Fatalf(
			"mark committing failed: %v",
			err,
		)
	}

	journal, err := Commit(
		tx.Staging,
		tx.Root,
		tx.Backup,
	)
	if err != nil {
		t.Fatalf(
			"commit failed: %v",
			err,
		)
	}

	if err := writeJournal(
		journalPath(workspace.Root),
		journal,
	); err != nil {
		t.Fatalf(
			"write journal failed: %v",
			err,
		)
	}

	// Simulate crash after filesystem modification.
	assertFileContent(
		t,
		filepath.Join(
			root,
			"usr",
			"bin",
			"cakra",
		),
		"cakra",
	)

	manager := NewRecoveryManager(parent)

	if err := manager.Recover(); err != nil {
		t.Fatalf(
			"recovery failed: %v",
			err,
		)
	}

	if _, err := os.Stat(
		filepath.Join(
			root,
			"usr",
			"bin",
			"cakra",
		),
	); !os.IsNotExist(err) {
		t.Fatal(
			"created file remained after crash recovery",
		)
	}
}

func TestRecoveryAfterCrashMultipleTransactions(t *testing.T) {
	parent := t.TempDir()

	workspace1, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace 1 failed: %v",
			err,
		)
	}

	workspace2, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace 2 failed: %v",
			err,
		)
	}

	if err := workspace1.Cleanup(); err != nil {
		t.Fatalf(
			"cleanup workspace 1 failed: %v",
			err,
		)
	}

	if err := writeState(
		workspace2.Root,
		StateCommitting,
	); err != nil {
		t.Fatalf(
			"write committing state failed: %v",
			err,
		)
	}

	// workspace2 deliberately has no journal.
	// Recovery must report this instead of silently
	// pretending the transaction was recovered.

	manager := NewRecoveryManager(parent)

	if err := manager.Recover(); err == nil {
		t.Fatal(
			"expected recovery error for missing journal",
		)
	}

	if _, err := os.Stat(
		workspace2.Root,
	); err != nil {
		t.Fatalf(
			"workspace disappeared unexpectedly: %v",
			err,
		)
	}
}
