package transaction

import (
	"fmt"
	"os"
	"path/filepath"
)

type Transaction struct {
	Workspace *Workspace
	Staging   string
	Root      string
	Backup    string

	Journal *Journal
}

func New(
	staging string,
	root string,
	backup string,
) *Transaction {
	return &Transaction{
		Staging: staging,
		Root:    root,
		Backup:  backup,
	}
}

func NewFromWorkspace(
	workspace *Workspace,
	root string,
) *Transaction {
	if workspace == nil {
		return &Transaction{
			Root: root,
		}
	}

	return &Transaction{
		Workspace: workspace,
		Staging:   workspace.Staging,
		Root:      root,
		Backup:    workspace.Backup,
	}
}

func (tx *Transaction) Commit() error {
	journal, err := Commit(
		tx.Staging,
		tx.Root,
		tx.Backup,
	)
	if err != nil {
		tx.Journal = journal
		return err
	}

	tx.Journal = journal

	return nil
}

func (tx *Transaction) Rollback() error {
	if tx.Journal == nil {
		return nil
	}

	return Rollback(tx.Journal)
}

func (tx *Transaction) Cleanup() error {
	if tx.Workspace != nil {
		return tx.Workspace.Cleanup()
	}
	if err := validateCleanupPath(
		tx.Staging,
	); err != nil {
		return fmt.Errorf(
			"invalid staging cleanup path: %w",
			err,
		)
	}

	if err := validateCleanupPath(
		tx.Backup,
	); err != nil {
		return fmt.Errorf(
			"invalid backup cleanup path: %w",
			err,
		)
	}

	var firstErr error

	if tx.Staging != "" {
		if err := os.RemoveAll(
			tx.Staging,
		); err != nil {
			firstErr = fmt.Errorf(
				"cleanup staging: %w",
				err,
			)
		}
	}

	if tx.Backup != "" {
		if err := os.RemoveAll(
			tx.Backup,
		); err != nil && firstErr == nil {
			firstErr = fmt.Errorf(
				"cleanup backup: %w",
				err,
			)
		}
	}

	return firstErr
}

/*func (tx *Transaction) Execute() error {
	if err := tx.Commit(); err != nil {
		rollbackErr := tx.Rollback()

		cleanupErr := tx.Cleanup()

		if rollbackErr != nil {
			return fmt.Errorf(
				"commit failed: %v; rollback failed: %w",
				err,
				rollbackErr,
			)
		}

		if cleanupErr != nil {
			return fmt.Errorf(
				"commit failed: %v; cleanup failed: %w",
				err,
				cleanupErr,
			)
		}

		return fmt.Errorf(
			"transaction commit failed: %w",
			err,
		)
	}

	if err := tx.Cleanup(); err != nil {
		return fmt.Errorf(
			"transaction cleanup failed: %w",
			err,
		)
	}

	return nil
}
*/

func validateCleanupPath(
	path string,
) error {
	if path == "" {
		return nil
	}

	clean := filepath.Clean(path)

	// Never allow filesystem root.
	if clean == string(filepath.Separator) {
		return fmt.Errorf(
			"refusing to remove filesystem root",
		)
	}

	// Never allow current directory.
	if clean == "." {
		return fmt.Errorf(
			"refusing to remove current directory",
		)
	}

	// Never allow parent traversal as a cleanup target.
	if clean == ".." ||
		filepath.IsAbs(clean) == false &&
			filepath.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf(
			"refusing unsafe cleanup path: %s",
			path,
		)
	}

	return nil
}
