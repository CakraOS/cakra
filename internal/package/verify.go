package packagefmt

import (
	"fmt"
	"os"
)

func VerifyGPK(
	path string,
	publicKeyPath string,
) error {
	gpk, err := ReadGPK(path)
	if err != nil {
		return err
	}

	if gpk.Metadata.Checksums == nil {
		return fmt.Errorf("package has no checksums")
	}

	// Verify payload
	tmpPayload, err := os.CreateTemp(
		"",
		"cakra-payload-*",
	)
	if err != nil {
		return err
	}

	payloadPath := tmpPayload.Name()
	defer os.Remove(payloadPath)

	if _, err := tmpPayload.Write(gpk.Payload); err != nil {
		tmpPayload.Close()
		return err
	}

	if err := tmpPayload.Close(); err != nil {
		return err
	}

	actualPayload, err := SHA256File(payloadPath)
	if err != nil {
		return err
	}

	if actualPayload != gpk.Metadata.Checksums.Payload {
		return fmt.Errorf(
			"payload checksum mismatch: expected %s, got %s",
			gpk.Metadata.Checksums.Payload,
			actualPayload,
		)
	}

	// Reconstruct manifest
	manifestData := []byte{}

	for _, file := range gpk.Manifest.Files {
		manifestData = append(
			manifestData,
			[]byte(file+"\n")...,
		)
	}

	tmpManifest, err := os.CreateTemp(
		"",
		"cakra-manifest-*",
	)
	if err != nil {
		return err
	}

	manifestPath := tmpManifest.Name()
	defer os.Remove(manifestPath)

	if _, err := tmpManifest.Write(manifestData); err != nil {
		tmpManifest.Close()
		return err
	}

	if err := tmpManifest.Close(); err != nil {
		return err
	}

	actualManifest, err := SHA256File(manifestPath)
	if err != nil {
		return err
	}

	if actualManifest != gpk.Metadata.Checksums.Manifest {
		return fmt.Errorf(
			"manifest checksum mismatch: expected %s, got %s",
			gpk.Metadata.Checksums.Manifest,
			actualManifest,
		)
	}

	// Verify signature
	if gpk.Metadata.Signature == nil {
		return fmt.Errorf("package has no signature")
	}

	publicKey, err := LoadPublicKey(publicKeyPath)
	if err != nil {
		return fmt.Errorf(
			"load public key: %w",
			err,
		)
	}

	if KeyID(publicKey) != gpk.Metadata.Signature.KeyID {
		return fmt.Errorf(
			"unknown signing key: %s",
			gpk.Metadata.Signature.KeyID,
		)
	}

	signingData := SigningData(gpk.Metadata)

	if !VerifySignature(
		publicKey,
		signingData,
		gpk.Metadata.Signature.Value,
	) {
		return fmt.Errorf(
			"Ed25519 signature verification failed",
		)
	}

	return nil
}
