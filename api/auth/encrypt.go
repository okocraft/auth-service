package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/Siroshun09/serrors"
)

type Encrypter interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(cipherData []byte) ([]byte, error)
}

type aesEncrypter struct {
	aead cipher.AEAD
}

func NewAESEncrypter(key []byte) (Encrypter, error) {
	if len(key) != 32 {
		return nil, serrors.New("key length must be 32")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &aesEncrypter{aead: aead}, nil
}

func (e *aesEncrypter) Encrypt(data []byte) ([]byte, error) {
	nonce, err := e.generateNonce()
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	return e.aead.Seal(nonce, nonce, data, nil), nil
}

func (e *aesEncrypter) Decrypt(cipherData []byte) ([]byte, error) {
	nonceSize := e.aead.NonceSize()

	if len(cipherData) < nonceSize {
		return nil, serrors.New("ciphertext too short")
	}

	nonce, data := cipherData[:nonceSize], cipherData[nonceSize:]
	decrypted, err := e.aead.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	return decrypted, nil
}

func (e *aesEncrypter) generateNonce() ([]byte, error) {
	iv := make([]byte, e.aead.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, serrors.WithStackTrace(err)
	}
	return iv, nil
}
