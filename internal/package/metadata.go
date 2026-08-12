package packagefmt

import (
	"encoding/json"
	"os"
)

type Checksums struct {
	Payload  string `json:"payload"`
	Manifest string `json:"manifest"`
}

type Signature struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Value     string `json:"value"`
}

type Metadata struct {
	Format       int        `json:"format"`
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	Release      int        `json:"release"`
	Architecture string     `json:"architecture"`
	Checksums    *Checksums `json:"checksums,omitempty"`
	Signature    *Signature `json:"signature,omitempty"`
}

func (m *Metadata) Write(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}
