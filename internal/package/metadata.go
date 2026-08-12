package packagefmt

import (
	"encoding/json"
	"os"
)

type Metadata struct {
	Format       int    `json:"format"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Release      int    `json:"release"`
	Architecture string `json:"architecture"`
}

func (m *Metadata) Write(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	return os.WriteFile(path, data, 0644)
}
