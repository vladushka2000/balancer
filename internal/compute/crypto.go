package compute

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

func deriveKey(secret []byte) []byte {
	sum := sha256.Sum256(secret)
	return sum[:]
}

func encrypt(plain, secret []byte) ([]byte, error) {
	if len(secret) == 0 {
		return plain, nil
	}
	block, err := aes.NewCipher(deriveKey(secret))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

func decrypt(data, secret []byte) ([]byte, error) {
	if len(secret) == 0 {
		return data, nil
	}
	block, err := aes.NewCipher(deriveKey(secret))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
