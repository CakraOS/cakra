package transaction

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultTransactionRoot = "tmp/cakra/transactions"

type Workspace struct {
	Root    string
	Staging string
	Backup  string
	State   State
}

func NewWorkspace(
	parent string,
) (*Workspace, error) {
	if parent == "" {
		parent = defaultTransactionRoot
	}

	if err := validateWorkspaceParent(parent); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(
		parent,
		0o755,
	); err != nil {
		return nil, fmt.Errorf(
			"create transaction parent: %w",
			err,
		)
	}

	id := fmt.Sprintf(
		"cakra-txn-%d",
		time.Now().UnixNano(),
	)

	root := filepath.Join(
		parent,
		id,
	)

	staging := filepath.Join(
		root,
		"staging",
	)

	backup := filepath.Join(
		root,
		"backup",
	)

	if err := os.MkdirAll(
		staging,
		0o755,
	); err != nil {
		return nil, fmt.Errorf(
			"create staging: %w",
			err,
		)
	}

	if err := os.MkdirAll(
		backup,
		0o755,
	); err != nil {
		os.RemoveAll(root)

		return nil, fmt.Errorf(
			"create backup: %w",
			err,
		)
	}

	if err := writeState(
		root,
		StateCreated,
	); err != nil {
		os.RemoveAll(root)

		return nil, fmt.Errorf(
			"create transaction state: %w",
			err,
		)
	}

	return &Workspace{
		Root:    root,
		Staging: staging,
		Backup:  backup,
		State:   StateCreated,
	}, nil
}

func (w *Workspace) Cleanup() error {
	if w == nil || w.Root == "" {
		return nil
	}

	if err := validateWorkspaceRoot(
		w.Root,
	); err != nil {
		return err
	}

	if err := os.RemoveAll(
		w.Root,
	); err != nil {
		return fmt.Errorf(
			"remove transaction workspace: %w",
			err,
		)
	}

	return nil
}

func validateWorkspaceParent(
	path string,
) error {
	clean := filepath.Clean(path)

	if clean == "." {
		return fmt.Errorf(
			"invalid workspace parent: %s",
			path,
		)
	}

	if clean == string(filepath.Separator) {
		return fmt.Errorf(
			"refusing filesystem root as workspace parent",
		)
	}

	return nil
}

func validateWorkspaceRoot(
	path string,
) error {
	clean := filepath.Clean(path)

	if clean == "." ||
		clean == string(filepath.Separator) {
		return fmt.Errorf(
			"refusing unsafe workspace root: %s",
			path,
		)
	}

	base := filepath.Base(clean)

	if len(base) < len("cakra-txn-") ||
		base[:len("cakra-txn-")] != "cakra-txn-" {
		return fmt.Errorf(
			"invalid transaction workspace: %s",
			path,
		)
	}

	return nil
}
