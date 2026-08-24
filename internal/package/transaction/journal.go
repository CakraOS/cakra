package transaction

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type journalFile struct {
	CreatedFiles []string `json:"created_files"`
	CreatedDirs  []string `json:"created_dirs"`
	Backups      []Backup `json:"backups"`
}

func journalPath(workspaceRoot string) string {
	return filepath.Join(
		workspaceRoot,
		"journal.json",
	)
}

func writeJournal(
	path string,
	journal *Journal,
) error {
	if journal == nil {
		return fmt.Errorf("journal is nil")
	}

	data, err := json.MarshalIndent(
		journalFile{
			CreatedFiles: journal.CreatedFiles,
			CreatedDirs:  journal.CreatedDirs,
			Backups:      journal.Backups,
		},
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf(
			"marshal journal: %w",
			err,
		)
	}

	data = append(data, '\n')

	dir := filepath.Dir(path)

	if err := os.MkdirAll(
		dir,
		0o755,
	); err != nil {
		return fmt.Errorf(
			"create journal directory: %w",
			err,
		)
	}

	tmp, err := os.CreateTemp(
		dir,
		".journal-*.tmp",
	)
	if err != nil {
		return fmt.Errorf(
			"create temporary journal: %w",
			err,
		)
	}

	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()

		return fmt.Errorf(
			"write temporary journal: %w",
			err,
		)
	}

	if err := tmp.Sync(); err != nil {
		cleanup()

		return fmt.Errorf(
			"sync temporary journal: %w",
			err,
		)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)

		return fmt.Errorf(
			"close temporary journal: %w",
			err,
		)
	}

	if err := os.Rename(
		tmpPath,
		path,
	); err != nil {
		_ = os.Remove(tmpPath)

		return fmt.Errorf(
			"replace journal: %w",
			err,
		)
	}

	return nil
}

func loadJournal(
	path string,
) (*Journal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var stored journalFile

	if err := json.Unmarshal(
		data,
		&stored,
	); err != nil {
		return nil, fmt.Errorf(
			"decode journal: %w",
			err,
		)
	}

	return &Journal{
		CreatedFiles: stored.CreatedFiles,
		CreatedDirs:  stored.CreatedDirs,
		Backups:      stored.Backups,
	}, nil
}
