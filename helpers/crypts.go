package helpers

import (
	"bytes"
	"crypto"
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

func (base *EncryptionBase) SignData(message string) (signature string, err error) {
	hashedMessage := base.createHash([]byte(message))

	privateKeyBase := base.rsaPrivateKey
	if privateKeyBase == "" {
		privateKeyBase = os.Getenv("PRIVATE_KEY_ENCRYPT")
	}

	privateKeyMarshal, err := base64.RawURLEncoding.DecodeString(privateKeyBase)
	if err != nil {
		return "", fmt.Errorf("error decode base64 rsa private key, err := %s", err.Error())
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyMarshal)
	if err != nil {
		return "", fmt.Errorf("error parse rsa private key, err := %s", err.Error())
	}

	signatureResult, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, []byte(hashedMessage), nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign the message, err := %s", err.Error())
	}

	result := signatureResult
	if base.encodeUrlBase64 {
		result = []byte(base64.RawURLEncoding.EncodeToString(signatureResult))
	}

	return string(result), nil
}

func (base *EncryptionBase) VerifyData(message string, signature string) (isVerified bool, err error) {
	hashedMessage := base.createHash([]byte(message))

	publicKeyBase := base.rsaPublicKey
	if publicKeyBase == "" {
		publicKeyBase = os.Getenv("PUBLIC_KEY_ENCRYPT")
	}

	publicKeyMarshal, err := base64.RawURLEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return false, fmt.Errorf("error decode base64 rsa private key, err := %s", err.Error())
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyMarshal)
	if err != nil {
		return false, fmt.Errorf("error parse rsa private key, err := %s", err.Error())
	}

	err = rsa.VerifyPSS(publicKey, crypto.SHA256, []byte(hashedMessage), []byte(signature), nil)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func GenerateRandString(strSize int, randType string) string {
	var dictionary string
	switch randType {
	case "alphanum":
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "alpha":
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "number":
		dictionary = "0123456789"
	default:
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	randString := make([]byte, strSize)
	_, _ = rand.Read(randString)

	for k, v := range randString {
		randString[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(randString)
}
