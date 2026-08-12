package packagefmt

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Manifest struct {
	Files []string
}

func GenerateManifest(root string) (*Manifest, error) {
	manifest := &Manifest{}

	err := filepath.Walk(root, func(
		path string,
		info os.FileInfo,
		err error,
	) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		relative = filepath.ToSlash(relative)

		manifest.Files = append(
			manifest.Files,
			relative,
		)

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(manifest.Files)

	return manifest, nil
}

func (m *Manifest) Write(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, file := range m.Files {
		file = strings.TrimSpace(file)

		if file == "" {
			continue
		}

		if _, err := writer.WriteString(file + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}
