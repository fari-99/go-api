package crypts

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

func (base *EncryptionBase) SetRsaKey(privateKey string, publicKey string) *EncryptionBase {
	base.rsaPrivateKey = privateKey
	base.rsaPublicKey = publicKey
	return base
}

func (base *EncryptionBase) GenerateRSAKey() (privateKey string, publicKey string, err error) {
	randomness := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(randomness, bitSize)
	if err != nil {
		return "", "", fmt.Errorf("error generate RSA key, err := %s", err.Error())
	}

	keyPrivate := x509.MarshalPKCS1PrivateKey(key)
	keyPublic := x509.MarshalPKCS1PublicKey(&key.PublicKey)

	// Save the private key and public key to ENV
	base.rsaPrivateKey = base64.RawURLEncoding.EncodeToString(keyPrivate)
	base.rsaPublicKey = base64.RawURLEncoding.EncodeToString(keyPublic)
	return base.rsaPrivateKey, base.rsaPublicKey, nil
}

func (base *EncryptionBase) EncryptRSA(secretMessage []byte) ([]byte, error) {
	passphrase := base.passphrase
	encodeBase64 := base.encodeUrlBase64

	publicKeyBase := base.rsaPublicKey
	if publicKeyBase == "" {
		publicKeyBase = os.Getenv("PUBLIC_KEY_ENCRYPT")
	}

	publicKeyMarshal, err := base64.RawURLEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return nil, fmt.Errorf("error decode base64 rsa public key, err := %s", err.Error())
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyMarshal)
	if err != nil {
		return nil, fmt.Errorf("error parse rsa public key, err := %s", err.Error())
	}

	var randomness io.Reader
	if base.useRandom {
		randomness = rand.Reader
	} else {
		randomness = bytes.NewReader(base.randomSetup)
	}

	cipherText, err := rsa.EncryptOAEP(sha256.New(), randomness, publicKey, secretMessage, passphrase)

	result := cipherText
	if encodeBase64 {
		result = []byte(base64.RawURLEncoding.EncodeToString(cipherText))
	}

	return result, err
}

func (base *EncryptionBase) DecryptRSA(secretMessage string) ([]byte, error) {
	passphrase := base.passphrase
	encodeBase64 := base.encodeUrlBase64

	privateKeyBase := base.rsaPrivateKey
	if privateKeyBase == "" {
		privateKeyBase = os.Getenv("PRIVATE_KEY_ENCRYPT")
	}

	privateKeyMarshal, err := base64.RawURLEncoding.DecodeString(privateKeyBase)
	if err != nil {
		return nil, fmt.Errorf("error decode base64 rsa private key, err := %s", err.Error())
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyMarshal)
	if err != nil {
		return nil, fmt.Errorf("error parse rsa private key, err := %s", err.Error())
	}

	var randomness io.Reader
	if base.useRandom {
		randomness = rand.Reader
	} else {
		randomness = bytes.NewReader(base.randomSetup)
	}

	var message []byte
	if encodeBase64 {
		message, err = base64.RawURLEncoding.DecodeString(secretMessage)
		if err != nil {
			return nil, fmt.Errorf("message is not base64 encoded, err := %s", err.Error())
		}
	}

	result, err := rsa.DecryptOAEP(sha256.New(), randomness, privateKey, message, passphrase)
	return result, err
}
