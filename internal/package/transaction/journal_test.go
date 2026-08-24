package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJournalPersistence(t *testing.T) {
	root := t.TempDir()

	path := filepath.Join(
		root,
		"journal.json",
	)

	journal := &Journal{
		Path: path,
		CreatedFiles: []string{
			"/root/usr/bin/cakra",
		},
		CreatedDirs: []string{
			"/root/usr/bin",
		},
		Backups: []Backup{
			{
				Original: "/root/usr/bin/hello",
				Backup:   "/backup/hello",
			},
		},
	}

	if err := writeJournal(
		path,
		journal,
	); err != nil {
		t.Fatalf(
			"write journal failed: %v",
			err,
		)
	}

	loaded, err := loadJournal(path)
	if err != nil {
		t.Fatalf(
			"load journal failed: %v",
			err,
		)
	}

	if len(loaded.CreatedFiles) != 1 {
		t.Fatalf(
			"expected 1 created file, got %d",
			len(loaded.CreatedFiles),
		)
	}

	if loaded.CreatedFiles[0] != journal.CreatedFiles[0] {
		t.Fatalf(
			"created file mismatch: %s != %s",
			loaded.CreatedFiles[0],
			journal.CreatedFiles[0],
		)
	}

	if len(loaded.CreatedDirs) != 1 {
		t.Fatalf(
			"expected 1 created directory, got %d",
			len(loaded.CreatedDirs),
		)
	}

	if len(loaded.Backups) != 1 {
		t.Fatalf(
			"expected 1 backup, got %d",
			len(loaded.Backups),
		)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf(
			"journal does not exist: %v",
			err,
		)
	}
}

func TestJournalWriteIsValidJSON(t *testing.T) {
	root := t.TempDir()

	path := filepath.Join(
		root,
		"journal.json",
	)

	journal := &Journal{
		Path: path,
		CreatedFiles: []string{
			"/root/a",
			"/root/b",
		},
	}

	if err := writeJournal(
		path,
		journal,
	); err != nil {
		t.Fatalf(
			"first journal write failed: %v",
			err,
		)
	}

	journal.CreatedFiles = append(
		journal.CreatedFiles,
		"/root/c",
	)

	if err := writeJournal(
		path,
		journal,
	); err != nil {
		t.Fatalf(
			"second journal write failed: %v",
			err,
		)
	}

	loaded, err := loadJournal(path)
	if err != nil {
		t.Fatalf(
			"journal became invalid after replacement: %v",
			err,
		)
	}

	if len(loaded.CreatedFiles) != 3 {
		t.Fatalf(
			"expected 3 created files, got %d",
			len(loaded.CreatedFiles),
		)
	}
}
