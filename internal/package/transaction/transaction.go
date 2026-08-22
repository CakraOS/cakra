package transaction

import (
	"fmt"
	"os"
	"path/filepath"
)

type Transaction struct {
	ID      string
	Staging string
}

func New(base string) (*Transaction, error) {
	id, err := os.MkdirTemp(
		base,
		"cakra-txn-*",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create transaction: %w",
			err,
		)
	}

	return &Transaction{
		ID:      filepath.Base(id),
		Staging: id,
	}, nil
}

func (t *Transaction) Root() string {
	return filepath.Join(
		t.Staging,
		"rootfs",
	)
}

func (t *Transaction) Cleanup() error {
	return os.RemoveAll(t.Staging)
}
