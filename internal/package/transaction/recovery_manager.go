package transaction

import "fmt"

type RecoveryManager struct {
	Parent string
}

func NewRecoveryManager(parent string) *RecoveryManager {
	if parent == "" {
		parent = defaultTransactionRoot
	}

	return &RecoveryManager{
		Parent: parent,
	}
}

func (rm *RecoveryManager) FindIncomplete() ([]string, error) {
	if rm == nil {
		return nil, fmt.Errorf("recovery manager is nil")
	}

	return FindIncompleteTransactions(rm.Parent)
}

func (rm *RecoveryManager) Recover() error {
	if rm == nil {
		return fmt.Errorf("recovery manager is nil")
	}

	transactions, err := rm.FindIncomplete()
	if err != nil {
		return fmt.Errorf(
			"find incomplete transactions: %w",
			err,
		)
	}

	var firstErr error

	for _, path := range transactions {
		if err := RecoverWorkspace(path); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf(
					"recover transaction %s: %w",
					path,
					err,
				)
			}
		}
	}

	return firstErr
}
