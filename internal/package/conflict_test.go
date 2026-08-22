package packagefmt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CakraOS/cakra/internal/package/db"
)

func TestCheckConflicts(t *testing.T) {
	tmp := t.TempDir()

	database := db.New(tmp)

	hello := db.Package{
		Name:         "hello",
		Version:      "0.1.0",
		Release:      1,
		Architecture: "aarch64",
		Files: []string{
			"usr/bin/hello",
		},
	}

	if err := database.Save(hello); err != nil {
		t.Fatal(err)
	}

	conflicting := db.Package{
		Name:         "evil",
		Version:      "1.0.0",
		Release:      1,
		Architecture: "aarch64",
		Files: []string{
			"usr/bin/hello",
		},
	}

	err := CheckConflicts(
		database,
		conflicting,
	)

	if err == nil {
		t.Fatal(
			"expected file conflict",
		)
	}

	_ = os.RemoveAll(
		filepath.Join(tmp, "packages"),
	)
}
