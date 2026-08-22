package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Package struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Release      int      `json:"release"`
	Architecture string   `json:"architecture"`
	Files        []string `json:"files"`
}

type Database struct {
	Root string
}

func New(root string) *Database {
	return &Database{
		Root: root,
	}
}

func (db *Database) packageDir(name string) string {
	return filepath.Join(
		db.Root,
		"packages",
		name,
	)
}

func (db *Database) Save(pkg Package) error {
	dir := db.packageDir(pkg.Name)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(
		pkg,
		"",
		"  ",
	)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	return os.WriteFile(
		filepath.Join(dir, "metadata.json"),
		data,
		0o644,
	)
}

func (db *Database) Load(name string) (*Package, error) {
	path := filepath.Join(
		db.packageDir(name),
		"metadata.json",
	)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg Package

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf(
			"invalid package database: %w",
			err,
		)
	}

	return &pkg, nil
}

func (db *Database) List() ([]Package, error) {
	dir := filepath.Join(
		db.Root,
		"packages",
	)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Package{}, nil
		}

		return nil, err
	}

	var packages []Package

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pkg, err := db.Load(entry.Name())
		if err != nil {
			return nil, err
		}

		packages = append(packages, *pkg)
	}

	return packages, nil
}

func (db *Database) Owner(
	file string,
) (*Package, error) {
	packages, err := db.List()
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		for _, owned := range pkg.Files {
			if owned == file {
				p := pkg
				return &p, nil
			}
		}
	}

	return nil, nil
}
func (db *Database) Remove(name string) error {
	dir := db.packageDir(name)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"package not installed: %s",
				name,
			)
		}

		return err
	}

	return os.RemoveAll(dir)
}
