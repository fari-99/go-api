package crypts

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

type EncryptionBase struct {
	passphrase      []byte
	encodeUrlBase64 bool

	useRandom   bool
	randomSetup []byte

	rsaPrivateKey string // encode to base64 raw url encoding
	rsaPublicKey  string // encode to base64 raw url encoding
}

func NewEncryptionBase() *EncryptionBase {
	base := &EncryptionBase{
		encodeUrlBase64: true,
		useRandom:       true,
	}

	return base
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

func (base *EncryptionBase) SetEncodeBase64(useEncode64 bool) *EncryptionBase {
	base.encodeUrlBase64 = useEncode64
	return base
}

func (base *EncryptionBase) createHash(passphrase []byte) string {
	hash := sha256.New()
	hash.Write(passphrase)
	return string(hash.Sum(nil))
}

func (base *EncryptionBase) Encrypt(data []byte) ([]byte, error) {
	blockKey, err := aes.NewCipher([]byte(base.createHash(base.passphrase)))
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
	blockKey, err := aes.NewCipher([]byte(base.createHash(base.passphrase)))
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