package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
)

var key []byte

func InitCrypto() {
	k := os.Getenv("CRYPTO_KEY")
	if len(k) != 32 {
		panic("CRYPTO_KEY must be 32 bytes")
	}
	key = []byte(k)
}

func Encrypt(text string) (string, error) {
	block, _ := aes.NewCipher(key)
	b := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], b)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encoded string) (string, error) {
	ciphertext, _ := base64.StdEncoding.DecodeString(encoded)
	block, _ := aes.NewCipher(key)
	if len(ciphertext) < aes.BlockSize {
		return "", nil
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}
