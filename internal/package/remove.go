package packagefmt

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CakraOS/cakra/internal/package/db"
)

func Remove(
	name string,
	root string,
	dbRoot string,
) error {
	database := db.New(dbRoot)

	pkg, err := database.Load(name)
	if err != nil {
		return fmt.Errorf(
			"load package: %w",
			err,
		)
	}

	for _, file := range pkg.Files {
		owner, err := database.Owner(file)
		if err != nil {
			return fmt.Errorf(
				"check ownership of %s: %w",
				file,
				err,
			)
		}

		if owner == nil {
			continue
		}

		if owner.Name != name {
			return fmt.Errorf(
				"refusing to remove %s: owned by %s",
				file,
				owner.Name,
			)
		}
	}

	for _, file := range pkg.Files {
		target := filepath.Join(root, file)

		if err := os.Remove(target); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return fmt.Errorf(
				"remove %s: %w",
				file,
				err,
			)
		}
	}

	if err := database.Remove(name); err != nil {
		return fmt.Errorf(
			"remove package database: %w",
			err,
		)
	}

	return nil
}
