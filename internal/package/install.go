package packagefmt

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CakraOS/cakra/internal/package/db"

	"github.com/CakraOS/cakra/internal/package/transaction"
)

func Install(
	gpkPath string,
	root string,
	publicKey string,
	dbRoot string,
) error {
	database := db.New(dbRoot)

	return installPackage(
		gpkPath,
		root,
		publicKey,
		database,
		database.Save,
	)
}

func installPackage(
	gpkPath string,
	root string,
	publicKey string,
	database *db.Database,
	save func(db.Package) error,
) error {
	if err := VerifyGPK(
		gpkPath,
		publicKey,
	); err != nil {
		return fmt.Errorf(
			"verification failed: %w",
			err,
		)
	}

	gpk, err := ReadGPK(gpkPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(
		root,
		0o755,
	); err != nil {
		return err
	}
	//database := db.New(dbRoot)
	installed := db.Package{
		Name:         gpk.Metadata.Name,
		Version:      gpk.Metadata.Version,
		Release:      gpk.Metadata.Release,
		Architecture: gpk.Metadata.Architecture,
		Files:        append([]string(nil), gpk.Manifest.Files...),
	}

	if err := CheckConflicts(
		database,
		installed,
	); err != nil {
		return fmt.Errorf(
			"package conflict: %w",
			err,
		)
	}
	/*
		txnBase := filepath.Join(
				os.TempDir(),
				"cakra",
			)

			if err := os.MkdirAll(
				txnBase,
				0o755,
			); err != nil {
				return fmt.Errorf(
					"create transaction directory: %w",
					err,
				)
			}

			txn, err := transaction.New(txnBase)
			if err != nil {
				return err
			}

			defer txn.Cleanup()
			if err := ExtractPayload(
				gpk.Payload,
				txn.Root(),
			); err != nil {
				return fmt.Errorf(
					"transaction extract payload: %w",
					err,
				)
			}

			/*installed := db.Package{
				Name:         gpk.Metadata.Name,
				Version:      gpk.Metadata.Version,
				Release:      gpk.Metadata.Release,
				Architecture: gpk.Metadata.Architecture,
				Files:        append([]string(nil), gpk.Manifest.Files...),
			}
	*/
	/*
		if err := transaction.Commit(
			txn.Root(),
			root,
		); err != nil {
			return fmt.Errorf(
				"transaction commit: %w",
				err,
			)
		}

		if err := database.Save(installed); err != nil {
			return fmt.Errorf(
				"save package database: %w",
				err,
			)
		}
	*/
	if err := RecoverTransactions(""); err != nil {
		return fmt.Errorf(
			"recover incomplete transactions: %w",
			err,
		)
	}
	txnBase := filepath.Join(
		"tmp",
		"cakra",
		"transactions",
	)

	workspace, err := transaction.NewWorkspace(
		txnBase,
	)
	if err != nil {
		return fmt.Errorf(
			"create transaction workspace: %w",
			err,
		)
	}

	tx := transaction.NewFromWorkspace(
		workspace,
		root,
	)

	defer func() {
		if err := tx.Cleanup(); err != nil {
			// Cleanup failure must not replace
			// the original transaction error.
		}
	}()

	if err := ExtractPayload(
		gpk.Payload,
		tx.Staging,
	); err != nil {
		return fmt.Errorf(
			"transaction extract payload: %w",
			err,
		)
	}
	if err := tx.MarkStaged(); err != nil {
		return fmt.Errorf(
			"mark transaction staged: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		rollbackErr := tx.Rollback()

		if rollbackErr != nil {
			return fmt.Errorf(
				"transaction commit failed: %v; rollback failed: %w",
				err,
				rollbackErr,
			)
		}

		return fmt.Errorf(
			"transaction commit: %w",
			err,
		)
	}

	if err := save(installed); err != nil {
		if markErr := tx.MarkFailed(); markErr != nil {
			return fmt.Errorf(
				"save package database failed: %v; mark transaction failed: %w",
				err,
				markErr,
			)
		}

		rollbackErr := tx.Rollback()

		if rollbackErr != nil {
			return fmt.Errorf(
				"save package database failed: %v; rollback failed: %w",
				err,
				rollbackErr,
			)
		}

		return fmt.Errorf(
			"save package database: %w",
			err,
		)
	}

	return nil
}
