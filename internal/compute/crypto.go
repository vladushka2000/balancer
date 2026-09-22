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

func newAEAD(secret []byte) (cipher.AEAD, error) {
	if len(secret) == 0 {
		return nil, nil
	}
	block, err := aes.NewCipher(deriveKey(secret))
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func encrypt(plain []byte, gcm cipher.AEAD) ([]byte, error) {
	if gcm == nil {
		return plain, nil
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

func decrypt(data []byte, gcm cipher.AEAD) ([]byte, error) {
	if gcm == nil {
		return data, nil
	}
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
