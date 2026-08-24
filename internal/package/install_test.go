package packagefmt

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CakraOS/cakra/internal/package/db"
	"github.com/klauspost/compress/zstd"
)

func TestInstallRollbackOnDatabaseFailure(t *testing.T) {
	root := t.TempDir()
	dbRoot := t.TempDir()
	fixtureDir := t.TempDir()

	// ------------------------------------------------------------
	// Prepare existing file.
	// This file will be overwritten by the package and must be
	// restored by Rollback().
	// ------------------------------------------------------------

	existingFile := filepath.Join(
		root,
		"usr",
		"bin",
		"existing",
	)

	writeInstallTestFile(
		t,
		existingFile,
		"original-content",
		0o755,
	)

	// ------------------------------------------------------------
	// Create signing key pair.
	// ------------------------------------------------------------

	privateKeyPath := filepath.Join(
		fixtureDir,
		"cakra-private.key",
	)

	publicKeyPath := filepath.Join(
		fixtureDir,
		"cakra-public.key",
	)

	if err := GenerateKeyPair(
		privateKeyPath,
		publicKeyPath,
	); err != nil {
		t.Fatalf(
			"generate key pair: %v",
			err,
		)
	}

	privateKey, err := LoadPrivateKey(
		privateKeyPath,
	)
	if err != nil {
		t.Fatalf(
			"load private key: %v",
			err,
		)
	}

	publicKey, err := LoadPublicKey(
		publicKeyPath,
	)
	if err != nil {
		t.Fatalf(
			"load public key: %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// Create real GPK fixture.
	// ------------------------------------------------------------

	payload := createInstallTestPayload(t)
	/*
		manifest := []string{
			"usr/bin/existing",
			"usr/bin/newfile",
		}
	*/
	payloadChecksum := sha256Hex(payload)

	manifestData := []byte(
		"usr/bin/existing\n" +
			"usr/bin/newfile\n",
	)

	manifestChecksum := sha256Hex(
		manifestData,
	)

	metadata := Metadata{
		Format:       1,
		Name:         "rollback-test",
		Version:      "1.0.0",
		Release:      1,
		Architecture: "arm64",
		Checksums: &Checksums{
			Payload:  payloadChecksum,
			Manifest: manifestChecksum,
		},
	}

	metadata.Signature = &Signature{
		Algorithm: "ed25519",
		KeyID:     KeyID(publicKey),
		Value: SignData(
			privateKey,
			SigningData(metadata),
		),
	}

	gpkPath := filepath.Join(
		fixtureDir,
		"rollback-test.gpk",
	)

	writeInstallTestGPK(
		t,
		gpkPath,
		metadata,
		manifestData,
		payload,
	)

	// ------------------------------------------------------------
	// Verify fixture itself before testing Install().
	// ------------------------------------------------------------

	if err := VerifyGPK(
		gpkPath,
		publicKeyPath,
	); err != nil {
		t.Fatalf(
			"test GPK verification failed: %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// Database.
	// ------------------------------------------------------------

	database := db.New(dbRoot)

	// Make sure database starts empty.
	packages, err := database.List()
	if err != nil {
		t.Fatalf(
			"initial database list: %v",
			err,
		)
	}

	if len(packages) != 0 {
		t.Fatalf(
			"database is not empty: %d packages",
			len(packages),
		)
	}

	// ------------------------------------------------------------
	// Inject database failure.
	// ------------------------------------------------------------

	saveErr := errors.New(
		"injected database failure",
	)

	save := func(
		pkg db.Package,
	) error {
		return saveErr
	}

	err = installPackage(
		gpkPath,
		root,
		publicKeyPath,
		database,
		save,
	)

	if err == nil {
		t.Fatal(
			"install should fail when database save fails",
		)
	}

	if !errors.Is(err, saveErr) {
		t.Fatalf(
			"expected database failure, got: %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// Verify overwritten file was restored.
	// ------------------------------------------------------------

	data, err := os.ReadFile(
		existingFile,
	)
	if err != nil {
		t.Fatalf(
			"read restored file: %v",
			err,
		)
	}

	if string(data) != "original-content" {
		t.Fatalf(
			"existing file was not restored: %q",
			string(data),
		)
	}

	// ------------------------------------------------------------
	// Verify newly-created file was removed.
	// ------------------------------------------------------------

	newFile := filepath.Join(
		root,
		"usr",
		"bin",
		"newfile",
	)

	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Fatalf(
			"new file still exists after rollback",
		)
	}

	// ------------------------------------------------------------
	// Verify database remains empty.
	// ------------------------------------------------------------

	packages, err = database.List()
	if err != nil {
		t.Fatalf(
			"database list after rollback: %v",
			err,
		)
	}

	if len(packages) != 0 {
		t.Fatalf(
			"database changed after failed install: %d packages",
			len(packages),
		)
	}

	// ------------------------------------------------------------
	// Verify package metadata does not exist.
	// ------------------------------------------------------------

	if _, err := database.Load(
		"rollback-test",
	); !os.IsNotExist(err) {
		t.Fatalf(
			"package metadata exists after failed install",
		)
	}
}

func createInstallTestPayload(
	t *testing.T,
) []byte {
	t.Helper()

	var tarBuffer bytes.Buffer

	tarWriter := tar.NewWriter(
		&tarBuffer,
	)

	files := []struct {
		name string
		data string
		mode int64
	}{
		{
			name: "usr/bin/existing",
			data: "new-content",
			mode: 0o755,
		},
		{
			name: "usr/bin/newfile",
			data: "new-file-content",
			mode: 0o644,
		},
	}

	for _, file := range files {
		header := &tar.Header{
			Name: file.name,
			Mode: file.mode,
			Size: int64(len(file.data)),
		}

		if err := tarWriter.WriteHeader(
			header,
		); err != nil {
			t.Fatalf(
				"write tar header: %v",
				err,
			)
		}

		if _, err := tarWriter.Write(
			[]byte(file.data),
		); err != nil {
			t.Fatalf(
				"write tar data: %v",
				err,
			)
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatalf(
			"close tar writer: %v",
			err,
		)
	}

	var payload bytes.Buffer

	zstdWriter, err := zstd.NewWriter(
		&payload,
	)
	if err != nil {
		t.Fatalf(
			"create zstd writer: %v",
			err,
		)
	}

	if _, err := zstdWriter.Write(
		tarBuffer.Bytes(),
	); err != nil {
		t.Fatalf(
			"compress payload: %v",
			err,
		)
	}

	if err := zstdWriter.Close(); err != nil {
		t.Fatalf(
			"close zstd writer: %v",
			err,
		)
	}

	return payload.Bytes()
}

func writeInstallTestGPK(
	t *testing.T,
	path string,
	metadata Metadata,
	manifest []byte,
	payload []byte,
) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf(
			"create GPK: %v",
			err,
		)
	}
	defer file.Close()

	if _, err := file.Write(
		[]byte("GPK\x01"),
	); err != nil {
		t.Fatalf(
			"write GPK magic: %v",
			err,
		)
	}

	metadataData, err := json.MarshalIndent(
		metadata,
		"",
		"  ",
	)
	if err != nil {
		t.Fatalf(
			"marshal metadata: %v",
			err,
		)
	}

	metadataData = append(
		metadataData,
		'\n',
	)

	writeInstallTestSection(
		t,
		file,
		"metadata",
		metadataData,
	)

	writeInstallTestSection(
		t,
		file,
		"manifest",
		manifest,
	)

	writeInstallTestSection(
		t,
		file,
		"payload",
		payload,
	)
}

func writeInstallTestSection(
	t *testing.T,
	file *os.File,
	name string,
	data []byte,
) {
	t.Helper()

	if len(name) > 255 {
		t.Fatalf(
			"section name too long",
		)
	}

	nameLen := uint8(len(name))

	if err := binary.Write(
		file,
		binary.LittleEndian,
		nameLen,
	); err != nil {
		t.Fatalf(
			"write section name length: %v",
			err,
		)
	}

	if _, err := file.Write(
		[]byte(name),
	); err != nil {
		t.Fatalf(
			"write section name: %v",
			err,
		)
	}

	size := uint64(len(data))

	if err := binary.Write(
		file,
		binary.LittleEndian,
		size,
	); err != nil {
		t.Fatalf(
			"write section size: %v",
			err,
		)
	}

	if _, err := file.Write(data); err != nil {
		t.Fatalf(
			"write section data: %v",
			err,
		)
	}
}

func writeInstallTestFile(
	t *testing.T,
	path string,
	content string,
	mode os.FileMode,
) {
	t.Helper()

	if err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	); err != nil {
		t.Fatalf(
			"create test directory: %v",
			err,
		)
	}

	if err := os.WriteFile(
		path,
		[]byte(content),
		mode,
	); err != nil {
		t.Fatalf(
			"write test file: %v",
			err,
		)
	}
}

func sha256Hex(
	data []byte,
) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(
		sum[:],
	)
}
