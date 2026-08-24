package transaction

import "fmt"

/*
	func (tx *Transaction) Execute() error {
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
func (tx *Transaction) Execute() error {
	if tx.Workspace != nil &&
		tx.Workspace.State == StateCreated {
		if err := tx.MarkStaged(); err != nil {
			return fmt.Errorf(
				"mark transaction staged: %w",
				err,
			)
		}
	}
	if tx.Workspace != nil &&
		tx.Workspace.State != StateCreated &&
		tx.Workspace.State != StateStaged {
		return fmt.Errorf(
			"cannot execute transaction from state: %s",
			tx.Workspace.State,
		)
	}

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
