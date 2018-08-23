package byteutil

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"crypto/rand"
	"errors"
)

func Encrypt(source []byte, key []byte) ([]byte, error) {
	if len(key) > 32 {
		key = key[:32]
	} else {
		padding := 32 - len(key)
		for i := 0; i < padding; i ++ {
			key = append(key, ' ')
		}
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, source, nil), nil
}

func Decrypt(encryptedData []byte, key []byte) ([]byte, error) {
	if len(key) > 32 {
		key = key[:32]
	} else {
		padding := 32 - len(key)
		for i := 0; i < padding; i ++ {
			key = append(key, ' ')
		}
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("data too short")
	}

	nonce, data := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, data, nil)
}
