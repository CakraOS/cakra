package packagefmt

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CakraOS/cakra/internal/package/db"

	"github.com/CakraOS/cakra/internal/package/transaction"
)

/*
func Install(

	gpkPath string,
	root string,
	publicKey string,

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

		if err := ExtractPayload(
			gpk.Payload,
			root,
		); err != nil {
			return fmt.Errorf(
				"extract payload: %w",
				err,
			)
		}
		packageFiles := make([]string, len(gpk.Manifest.Files))

		copy(packageFiles, gpk.Manifest.Files)

		return nil
	}
*/
func Install(
	gpkPath string,
	root string,
	publicKey string,
	dbRoot string,
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
	database := db.New(dbRoot)
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

	return nil
}
