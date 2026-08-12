package packagefmt

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

var magic = [4]byte{'G', 'P', 'K', 0x01}

func writeSection(
	w io.Writer,
	name string,
	data []byte,
) error {
	if len(name) > 255 {
		return fmt.Errorf("section name too long")
	}

	if _, err := w.Write([]byte{byte(len(name))}); err != nil {
		return err
	}

	if _, err := w.Write([]byte(name)); err != nil {
		return err
	}

	if err := binary.Write(
		w,
		binary.LittleEndian,
		uint64(len(data)),
	); err != nil {
		return err
	}

	_, err := w.Write(data)

	return err
}

func createContainer(
	output string,
	metadataPath string,
	manifestPath string,
	payloadPath string,
) error {
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := out.Write(magic[:]); err != nil {
		return err
	}

	metadata, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	payload, err := os.ReadFile(payloadPath)
	if err != nil {
		return err
	}

	if err := writeSection(
		out,
		"metadata",
		metadata,
	); err != nil {
		return err
	}

	if err := writeSection(
		out,
		"manifest",
		manifest,
	); err != nil {
		return err
	}

	if err := writeSection(
		out,
		"payload",
		payload,
	); err != nil {
		return err
	}

	return nil
}
