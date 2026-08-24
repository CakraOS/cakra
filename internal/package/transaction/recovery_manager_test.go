package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecoveryManagerFindIncomplete(t *testing.T) {
	parent := t.TempDir()

	workspace, err := NewWorkspace(parent)
	if err != nil {
		t.Fatalf(
			"create workspace failed: %v",
			err,
		)
	}

	manager := NewRecoveryManager(parent)

	found, err := manager.FindIncomplete()
	if err != nil {
		t.Fatalf(
			"find incomplete transactions failed: %v",
			err,
		)
	}

	if len(found) != 1 {
		t.Fatalf(
			"expected 1 incomplete transaction, got %d",
			len(found),
		)
	}

	if found[0] != workspace.Root {
		t.Fatalf(
			"unexpected transaction path: %s",
			found[0],
		)
	}
}

func TestRecoveryManagerRecoverCreated(t *testing.T) {
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

	manager := NewRecoveryManager(parent)

	if err := manager.Recover(); err != nil {
		t.Fatalf(
			"recovery failed: %v",
			err,
		)
	}

	if _, err := os.Stat(
		workspace.Root,
	); !os.IsNotExist(err) {
		t.Fatal(
			"recovered workspace still exists",
		)
	}
}

func TestRecoveryManagerRecoverMultiple(t *testing.T) {
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

	manager := NewRecoveryManager(parent)

	if err := manager.Recover(); err != nil {
		t.Fatalf(
			"recovery failed: %v",
			err,
		)
	}

	for _, workspace := range []*Workspace{
		workspace1,
		workspace2,
	} {
		if _, err := os.Stat(
			workspace.Root,
		); !os.IsNotExist(err) {
			t.Fatalf(
				"workspace still exists: %s",
				workspace.Root,
			)
		}
	}
}

func TestRecoveryManagerEmptyParent(t *testing.T) {
	parent := t.TempDir()

	manager := NewRecoveryManager(parent)

	if err := manager.Recover(); err != nil {
		t.Fatalf(
			"empty recovery failed: %v",
			err,
		)
	}
}

func TestRecoveryManagerNil(t *testing.T) {
	var manager *RecoveryManager

	if _, err := manager.FindIncomplete(); err == nil {
		t.Fatal(
			"expected nil manager error",
		)
	}

	if err := manager.Recover(); err == nil {
		t.Fatal(
			"expected nil manager error",
		)
	}
}
