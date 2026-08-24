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

/*
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
*/
/*func (tx *Transaction) Commit() error {
	if tx.Workspace != nil {
		if tx.Workspace.State != StateStaged {
			return fmt.Errorf(
				"cannot commit transaction from state: %s",
				tx.Workspace.State,
			)
		}

		if err := tx.MarkCommitting(); err != nil {
			return err
		}
	}

	journal, err := Commit(
		tx.Staging,
		tx.Root,
		tx.Backup,
	)

	tx.Journal = journal

	if err != nil {
		return err
	}

	if tx.Workspace != nil {
		if err := tx.MarkCommitted(); err != nil {
			return fmt.Errorf(
				"mark transaction committed: %w",
				err,
			)
		}
	}

	return nil
}
*/
func (tx *Transaction) Commit() error {
	if tx.Workspace != nil {
		if tx.Workspace.State != StateStaged {
			return fmt.Errorf(
				"cannot commit transaction from state: %s",
				tx.Workspace.State,
			)
		}

		if err := tx.MarkCommitting(); err != nil {
			return err
		}
	}

	journal, err := Commit(
		tx.Staging,
		tx.Root,
		tx.Backup,
	)

	tx.Journal = journal

	if err != nil {
		if tx.Workspace != nil {
			if stateErr := tx.MarkFailed(); stateErr != nil {
				return fmt.Errorf(
					"commit failed: %v; mark failed: %w",
					err,
					stateErr,
				)
			}
		}

		return err
	}

	if tx.Workspace != nil {
		if err := tx.MarkCommitted(); err != nil {
			return fmt.Errorf(
				"mark transaction committed: %w",
				err,
			)
		}
	}

	return nil
}

/*
	func (tx *Transaction) MarkFailed() error {
		if tx.Workspace == nil {
			return fmt.Errorf(
				"transaction workspace is nil",
			)
		}

		if tx.Workspace.State != StateCommitting {
			return fmt.Errorf(
				"invalid state transition: %s -> %s",
				tx.Workspace.State,
				StateFailed,
			)
		}

		if err := writeState(
			tx.Workspace.Root,
			StateFailed,
		); err != nil {
			return fmt.Errorf(
				"write failed state: %w",
				err,
			)
		}

		tx.Workspace.State = StateFailed

		return nil
	}
*/
func (tx *Transaction) MarkFailed() error {
	if tx.Workspace == nil {
		return fmt.Errorf(
			"transaction workspace is nil",
		)
	}

	if tx.Workspace.State != StateCommitting &&
		tx.Workspace.State != StateCommitted {
		return fmt.Errorf(
			"invalid state transition: %s -> %s",
			tx.Workspace.State,
			StateFailed,
		)
	}

	if err := writeState(
		tx.Workspace.Root,
		StateFailed,
	); err != nil {
		return fmt.Errorf(
			"write failed state: %w",
			err,
		)
	}

	tx.Workspace.State = StateFailed

	return nil
}

func (tx *Transaction) MarkRollingBack() error {
	if tx.Workspace == nil {
		return fmt.Errorf(
			"transaction workspace is nil",
		)
	}

	if tx.Workspace.State != StateFailed {
		return fmt.Errorf(
			"invalid state transition: %s -> %s",
			tx.Workspace.State,
			StateRollingBack,
		)
	}

	if err := writeState(
		tx.Workspace.Root,
		StateRollingBack,
	); err != nil {
		return fmt.Errorf(
			"write rolling back state: %w",
			err,
		)
	}

	tx.Workspace.State = StateRollingBack

	return nil
}

func (tx *Transaction) MarkRolledBack() error {
	if tx.Workspace == nil {
		return fmt.Errorf(
			"transaction workspace is nil",
		)
	}

	if tx.Workspace.State != StateRollingBack {
		return fmt.Errorf(
			"invalid state transition: %s -> %s",
			tx.Workspace.State,
			StateRolledBack,
		)
	}

	if err := writeState(
		tx.Workspace.Root,
		StateRolledBack,
	); err != nil {
		return fmt.Errorf(
			"write rolled back state: %w",
			err,
		)
	}

	tx.Workspace.State = StateRolledBack

	return nil
}

/*
	func (tx *Transaction) Rollback() error {
		if tx.Journal == nil {
			return nil
		}

		return Rollback(tx.Journal)
	}
*/
func (tx *Transaction) Rollback() error {
	if tx.Journal == nil {
		return nil
	}

	if tx.Workspace != nil {
		if tx.Workspace.State != StateFailed {
			return fmt.Errorf(
				"cannot rollback transaction from state: %s",
				tx.Workspace.State,
			)
		}

		if err := tx.MarkRollingBack(); err != nil {
			return err
		}
	}

	if err := Rollback(tx.Journal); err != nil {
		return err
	}

	if tx.Workspace != nil {
		if err := tx.MarkRolledBack(); err != nil {
			return fmt.Errorf(
				"mark transaction rolled back: %w",
				err,
			)
		}
	}

	return nil
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

func (tx *Transaction) MarkStaged() error {
	if tx.Workspace == nil {
		return fmt.Errorf(
			"transaction workspace is nil",
		)
	}

	if tx.Workspace.State != StateCreated {
		return fmt.Errorf(
			"invalid state transition: %s -> %s",
			tx.Workspace.State,
			StateStaged,
		)
	}

	if err := writeState(
		tx.Workspace.Root,
		StateStaged,
	); err != nil {
		return fmt.Errorf(
			"write staged state: %w",
			err,
		)
	}

	tx.Workspace.State = StateStaged

	return nil
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

func (tx *Transaction) MarkCommitting() error {
	if tx.Workspace == nil {
		return fmt.Errorf(
			"transaction workspace is nil",
		)
	}

	if tx.Workspace.State != StateStaged {
		return fmt.Errorf(
			"invalid state transition: %s -> %s",
			tx.Workspace.State,
			StateCommitting,
		)
	}

	if err := writeState(
		tx.Workspace.Root,
		StateCommitting,
	); err != nil {
		return fmt.Errorf(
			"write committing state: %w",
			err,
		)
	}

	tx.Workspace.State = StateCommitting

	return nil
}

func (tx *Transaction) MarkCommitted() error {
	if tx.Workspace == nil {
		return fmt.Errorf(
			"transaction workspace is nil",
		)
	}

	if tx.Workspace.State != StateCommitting {
		return fmt.Errorf(
			"invalid state transition: %s -> %s",
			tx.Workspace.State,
			StateCommitted,
		)
	}

	if err := writeState(
		tx.Workspace.Root,
		StateCommitted,
	); err != nil {
		return fmt.Errorf(
			"write committed state: %w",
			err,
		)
	}

	tx.Workspace.State = StateCommitted

	return nil
}
