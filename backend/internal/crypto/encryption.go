package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	EncryptedPrefix = "encrypted:"
	MasterKeyEnv    = "OMNIPOLL_MASTER_KEY"
)

// Encryptor handles encryption/decryption of sensitive data
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new Encryptor using the master key from environment
func NewEncryptor() (*Encryptor, error) {
	masterKey := os.Getenv(MasterKeyEnv)
	if masterKey == "" {
		// Use a default key for development (NOT for production!)
		masterKey = "omnipoll-dev-key-change-in-prod"
	}

	// Derive a 32-byte key using SHA256
	hash := sha256.Sum256([]byte(masterKey))
	return &Encryptor{key: hash[:]}, nil
}

// NewEncryptorWithKey creates an Encryptor with a specific key
func NewEncryptorWithKey(masterKey string) *Encryptor {
	hash := sha256.Sum256([]byte(masterKey))
	return &Encryptor{key: hash[:]}
}

// Encrypt encrypts plaintext and returns base64-encoded ciphertext with prefix
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return EncryptedPrefix + encoded, nil
}

// Decrypt decrypts a prefixed encrypted string
func (e *Encryptor) Decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	// If not encrypted, return as-is
	if !strings.HasPrefix(encrypted, EncryptedPrefix) {
		return encrypted, nil
	}

	encoded := strings.TrimPrefix(encrypted, EncryptedPrefix)
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a string is encrypted
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, EncryptedPrefix)
}
