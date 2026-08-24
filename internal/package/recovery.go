package packagefmt

import (
	"fmt"

	"github.com/CakraOS/cakra/internal/package/transaction"
)

func RecoverTransactions(parent string) error {
	manager := transaction.NewRecoveryManager(parent)

	if err := manager.Recover(); err != nil {
		return fmt.Errorf(
			"recover package transactions: %w",
			err,
		)
	}

	return nil
}
