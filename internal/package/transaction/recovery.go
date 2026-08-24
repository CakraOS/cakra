package transaction

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindIncompleteTransactions(parent string) ([]string, error) {
	if parent == "" {
		parent = defaultTransactionRoot
	}

	if err := validateWorkspaceParent(parent); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, fmt.Errorf(
			"read transaction parent: %w",
			err,
		)
	}

	var incomplete []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if len(entry.Name()) < len("cakra-txn-") ||
			entry.Name()[:len("cakra-txn-")] != "cakra-txn-" {
			continue
		}

		path := filepath.Join(
			parent,
			entry.Name(),
		)

		state, err := readState(path)
		if err != nil {
			return nil, fmt.Errorf(
				"read transaction state %s: %w",
				path,
				err,
			)
		}

		switch state {
		case StateCreated,
			StateStaged,
			StateCommitting,
			StateFailed,
			StateRollingBack:

			incomplete = append(
				incomplete,
				path,
			)

		case StateCommitted,
			StateRolledBack:

			continue

		default:
			return nil, fmt.Errorf(
				"unknown transaction state %s: %s",
				state,
				path,
			)
		}
	}

	return incomplete, nil
}

func RecoverWorkspace(path string) error {
	if err := validateWorkspaceRoot(path); err != nil {
		return err
	}

	state, err := readState(path)
	if err != nil {
		return fmt.Errorf(
			"read transaction state: %w",
			err,
		)
	}

	switch state {

	case StateCreated, StateStaged:
		// The transaction has not committed anything
		// to the target root yet.
		return cleanupRecoveredWorkspace(path)

	case StateCommitted, StateRolledBack:
		// Nothing needs to be rolled back.
		return cleanupRecoveredWorkspace(path)

	case StateCommitting,
		StateFailed,
		StateRollingBack:

		journal, err := loadJournal(
			journalPath(path),
		)
		if err != nil {
			return fmt.Errorf(
				"load transaction journal: %w",
				err,
			)
		}

		if err := writeState(
			path,
			StateRollingBack,
		); err != nil {
			return fmt.Errorf(
				"mark transaction rolling back: %w",
				err,
			)
		}

		if err := Rollback(journal); err != nil {
			return fmt.Errorf(
				"rollback recovered transaction: %w",
				err,
			)
		}

		if err := writeState(
			path,
			StateRolledBack,
		); err != nil {
			return fmt.Errorf(
				"mark transaction rolled back: %w",
				err,
			)
		}

		return cleanupRecoveredWorkspace(path)

	default:
		return fmt.Errorf(
			"cannot recover transaction from state: %s",
			state,
		)
	}
}

func cleanupRecoveredWorkspace(path string) error {
	if err := validateWorkspaceRoot(path); err != nil {
		return err
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf(
			"cleanup recovered workspace: %w",
			err,
		)
	}

	return nil
}
