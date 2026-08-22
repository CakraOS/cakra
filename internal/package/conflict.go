package packagefmt

import (
	"fmt"

	"github.com/CakraOS/cakra/internal/package/db"
)

func CheckConflicts(
	database *db.Database,
	pkg db.Package,
) error {
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

		if owner.Name == pkg.Name {
			continue
		}

		return fmt.Errorf(
			"file conflict: %s is owned by %s",
			file,
			owner.Name,
		)
	}

	return nil
}
