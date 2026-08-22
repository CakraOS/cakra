package transaction

import "fmt"

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
