package packagefmt

import (
	"archive/tar"
	"crypto/ed25519"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func createPayload(root string, output string) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	return filepath.Walk(root, func(
		path string,
		info os.FileInfo,
		err error,
	) error {
		if err != nil {
			return err
		}

		if path == root {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		relative = filepath.ToSlash(relative)

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = relative

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(tw, src)

		return err
	})
}

/*
func BuildGPK(
	output string,
	root string,
	metadata Metadata,
) error {
	tmpDir, err := os.MkdirTemp("", "cakra-gpk-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	manifestPath := filepath.Join(tmpDir, "manifest")
	payloadPath := filepath.Join(tmpDir, "payload.tar")

	if err := metadata.Write(metadataPath); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	manifest, err := GenerateManifest(root)
	if err != nil {
		return fmt.Errorf("generate manifest: %w", err)
	}

	if err := manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	if err := createPayload(root, payloadPath); err != nil {
		return fmt.Errorf("create payload: %w", err)
	}

	return createContainer(
		output,
		metadataPath,
		manifestPath,
		payloadPath,
	)
}
*/

func BuildGPK(
	output string,
	root string,
	metadata Metadata,
) error {
	tmpDir, err := os.MkdirTemp("", "cakra-gpk-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	manifestPath := filepath.Join(tmpDir, "manifest")
	payloadTar := filepath.Join(tmpDir, "payload.tar")
	payloadPath := filepath.Join(tmpDir, "payload.tar.zst")

	manifest, err := GenerateManifest(root)
	if err != nil {
		return fmt.Errorf("generate manifest: %w", err)
	}

	if err := manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	manifestChecksum, err := SHA256File(manifestPath)
	if err != nil {
		return fmt.Errorf("checksum manifest: %w", err)
	}

	if err := createPayload(root, payloadTar); err != nil {
		return fmt.Errorf("create payload: %w", err)
	}
	if err := CompressZstd(payloadTar, payloadPath); err != nil {
		return fmt.Errorf("compress payload: %w", err)
	}

	payloadChecksum, err := SHA256File(payloadPath)
	if err != nil {
		return fmt.Errorf("checksum payload: %w", err)
	}

	metadata.Checksums = &Checksums{
		Payload:  payloadChecksum,
		Manifest: manifestChecksum,
	}
	privateKey, err := LoadPrivateKey("keys/cakra-private.key")
	if err != nil {
		return fmt.Errorf("load signing key: %w", err)
	}

	signingData := SigningData(metadata)

	signature := SignData(
		privateKey,
		signingData,
	)

	publicKey := privateKey.Public().(ed25519.PublicKey)

	metadata.Signature = &Signature{
		Algorithm: "Ed25519",
		KeyID:     KeyID(publicKey),
		Value:     signature,
	}

	if err := metadata.Write(metadataPath); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	return createContainer(
		output,
		metadataPath,
		manifestPath,
		payloadPath,
	)
}
