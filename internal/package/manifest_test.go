package packagefmt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateManifest(t *testing.T) {
	root := t.TempDir()

	file := filepath.Join(root, "usr", "bin", "hello")

	if err := os.MkdirAll(
		filepath.Dir(file),
		0755,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		file,
		[]byte("hello"),
		0755,
	); err != nil {
		t.Fatal(err)
	}

	manifest, err := GenerateManifest(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(manifest.Files) != 1 {
		t.Fatalf(
			"expected 1 file, got %d",
			len(manifest.Files),
		)
	}

	if manifest.Files[0] != "usr/bin/hello" {
		t.Fatalf(
			"unexpected manifest entry: %s",
			manifest.Files[0],
		)
	}
}
