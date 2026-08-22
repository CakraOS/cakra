package transaction

import (
	"fmt"
	"os"
	"path/filepath"
)

func Rollback(
	journal *Journal,
) error {
	if journal == nil {
		return nil
	}

	var firstErr error

	// Restore files that existed before the transaction.
	for i := len(journal.Backups) - 1; i >= 0; i-- {
		backup := journal.Backups[i]

		if err := restoreBackup(
			backup,
		); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Remove files that were created by the transaction.
	for i := len(journal.CreatedFiles) - 1; i >= 0; i-- {
		path := journal.CreatedFiles[i]

		if err := removePath(path); err != nil &&
			!os.IsNotExist(err) &&
			firstErr == nil {
			firstErr = err
		}
	}

	// Remove directories created by the transaction.
	// Reverse order is important because child directories
	// must disappear before their parents.
	for i := len(journal.CreatedDirs) - 1; i >= 0; i-- {
		path := journal.CreatedDirs[i]

		if err := os.Remove(path); err != nil &&
			!os.IsNotExist(err) &&
			firstErr == nil {
			firstErr = fmt.Errorf(
				"remove directory %s: %w",
				path,
				err,
			)
		}
	}

	return firstErr
}

func restoreBackup(
	backup Backup,
) error {
	if err := os.MkdirAll(
		filepath.Dir(backup.Original),
		0755,
	); err != nil {
		return fmt.Errorf(
			"create restore directory: %w",
			err,
		)
	}

	if err := os.Remove(
		backup.Original,
	); err != nil &&
		!os.IsNotExist(err) {
		return fmt.Errorf(
			"remove current file %s: %w",
			backup.Original,
			err,
		)
	}

	if err := copyRaw(
		backup.Backup,
		backup.Original,
	); err != nil {
		return fmt.Errorf(
			"restore %s: %w",
			backup.Original,
			err,
		)
	}

	return nil
}

func removePath(
	path string,
) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf(
			"remove %s: %w",
			path,
			err,
		)
	}

	return nil
}
