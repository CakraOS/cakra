package packagefmt

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

/*
type Signature struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Value     string `json:"value"`
}
*/

func GenerateKeyPair(
	privatePath string,
	publicPath string,
) error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	if err := os.WriteFile(
		privatePath,
		privateKey,
		0o600,
	); err != nil {
		return fmt.Errorf("write private key: %w", err)
	}

	if err := os.WriteFile(
		publicPath,
		publicKey,
		0o644,
	); err != nil {
		return fmt.Errorf("write public key: %w", err)
	}

	return nil
}

func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf(
			"invalid private key size: %d",
			len(data),
		)
	}

	return ed25519.PrivateKey(data), nil
}

func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf(
			"invalid public key size: %d",
			len(data),
		)
	}

	return ed25519.PublicKey(data), nil
}

func KeyID(publicKey ed25519.PublicKey) string {
	hash := sha256.Sum256(publicKey)

	return hex.EncodeToString(hash[:8])
}

func SignData(
	privateKey ed25519.PrivateKey,
	data []byte,
) string {
	signature := ed25519.Sign(
		privateKey,
		data,
	)

	return hex.EncodeToString(signature)
}

func VerifySignature(
	publicKey ed25519.PublicKey,
	data []byte,
	signatureHex string,
) bool {
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	return ed25519.Verify(
		publicKey,
		data,
		signature,
	)
}

func SigningData(m Metadata) []byte {
	payloadChecksum := ""
	manifestChecksum := ""

	if m.Checksums != nil {
		payloadChecksum = m.Checksums.Payload
		manifestChecksum = m.Checksums.Manifest
	}

	data := fmt.Sprintf(
		"cakra-gpk:%d\n"+
			"name:%s\n"+
			"version:%s\n"+
			"release:%d\n"+
			"architecture:%s\n"+
			"payload-sha256:%s\n"+
			"manifest-sha256:%s\n",
		m.Format,
		m.Name,
		m.Version,
		m.Release,
		m.Architecture,
		payloadChecksum,
		manifestChecksum,
	)

	return []byte(data)
}
