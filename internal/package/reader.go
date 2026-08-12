package packagefmt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Section struct {
	Name string
	Data []byte
}

type GPK struct {
	Metadata Metadata
	Manifest Manifest
	Payload  []byte
}

func readSection(r io.Reader) (*Section, error) {
	var nameLen uint8

	if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
		return nil, err
	}

	nameBytes := make([]byte, nameLen)

	if _, err := io.ReadFull(r, nameBytes); err != nil {
		return nil, err
	}

	var size uint64

	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	if size > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("section too large")
	}

	data := make([]byte, int(size))

	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return &Section{
		Name: string(nameBytes),
		Data: data,
	}, nil
}

func ReadGPK(path string) (*GPK, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var header [4]byte

	if _, err := io.ReadFull(file, header[:]); err != nil {
		return nil, err
	}

	if header != magic {
		return nil, fmt.Errorf("invalid GPK magic")
	}

	result := &GPK{}

	for {
		section, err := readSection(file)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch section.Name {
		case "metadata":
			if err := jsonUnmarshal(section.Data, &result.Metadata); err != nil {
				return nil, fmt.Errorf("invalid metadata: %w", err)
			}

		case "manifest":
			result.Manifest.Files = parseManifest(section.Data)

		case "payload":
			result.Payload = section.Data

		default:
			return nil, fmt.Errorf(
				"unknown section: %s",
				section.Name,
			)
		}
	}

	return result, nil
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func parseManifest(data []byte) []string {
	lines := strings.Split(string(data), "\n")

	files := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line != "" {
			files = append(files, line)
		}
	}

	return files
}
