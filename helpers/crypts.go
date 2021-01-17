package helpers

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type EncryptionBase struct {
	passphrase      []byte
	encodeUrlBase64 bool

	useRandom   bool
	randomSetup []byte
}

func NewEncryptionBase() *EncryptionBase {
	base := &EncryptionBase{
		encodeUrlBase64: true,
		useRandom:       true,
	}

	return base
}

func (base *EncryptionBase) GenerateRSAKey() error {
	randomness := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(randomness, bitSize)
	if err != nil {
		return fmt.Errorf("error generate RSA key, err := %s", err.Error())
	}

	privateKey := x509.MarshalPKCS1PrivateKey(key)
	publicKey := x509.MarshalPKCS1PublicKey(&key.PublicKey)

	// Save the private key and public key to ENV
	log.Printf(base64.RawURLEncoding.EncodeToString(privateKey))
	log.Printf(base64.RawURLEncoding.EncodeToString(publicKey))
	return nil
}

func (base *EncryptionBase) SetPassphrase(passphrase string) *EncryptionBase {
	if passphrase != "" {
		base.passphrase = []byte(passphrase)
	}

	return base
}

func (base *EncryptionBase) SetUseRandomness(useRandomness bool, keyRandom string) *EncryptionBase {
	base.useRandom = useRandomness

	if !useRandomness {
		if keyRandom != "" {
			log.Printf(keyRandom)
			base.randomSetup = []byte(keyRandom)
		} else {
			// randomness used for security data, if you are not using it. it still need some key random,
			// so you must set it.
			panic("need to set random key when use random is false.")
		}
	}

	return base
}

func (base *EncryptionBase) createHash() string {
	passphrase := base.passphrase

	hash := sha256.New()
	hash.Write(passphrase)
	return string(hash.Sum(nil))
}

func (base *EncryptionBase) EncryptRSA(secretMessage []byte) ([]byte, error) {
	passphrase := base.passphrase
	encodeBase64 := base.encodeUrlBase64

	publicKeyMarshal, err := base64.RawURLEncoding.DecodeString(os.Getenv("PUBLIC_KEY_ENCRYPT"))
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

	privateKeyMarshal, err := base64.RawURLEncoding.DecodeString(os.Getenv("PRIVATE_KEY_ENCRYPT"))
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

func (base *EncryptionBase) Encrypt(data []byte) ([]byte, error) {
	blockKey, err := aes.NewCipher([]byte(base.createHash()))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm, err := %s", err.Error())
	}

	var randomness io.Reader
	if base.useRandom {
		randomness = rand.Reader
	} else {
		randomness = bytes.NewReader(base.randomSetup)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(randomness, nonce); err != nil {
		return nil, fmt.Errorf("failed to create nonce, err := %s", err.Error())
	}

	cipherText := gcm.Seal(nonce, nonce, data, nil)

	result := cipherText
	if base.encodeUrlBase64 {
		result = []byte(base64.RawURLEncoding.EncodeToString(cipherText))
	}

	return result, nil
}

func (base *EncryptionBase) Decrypt(data []byte) ([]byte, error) {
	blockKey, err := aes.NewCipher([]byte(base.createHash()))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockKey)
	if err != nil {
		return nil, err
	}

	if base.encodeUrlBase64 {
		data, err = base64.RawURLEncoding.DecodeString(string(data))
		if err != nil {
			return nil, err
		}
	}

	nonceSize := gcm.NonceSize()
	nonce, cipherText := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
