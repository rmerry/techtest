// Package encryption is a thin layer around a handful of the standard
// libraries symmetric encryption algorithms. It makes using them slightly
// easier by providing a simple interface for encrypting and decrypting
// messages given just an algorithm name, key and body.
//
// It is understood that this wrapper is primitive and rudimentary and should
// not be used in production, it is simply for illustrative purposes.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	// ErrCipherCreation indicates a problem creating a new cipher object, an
	// invalid key is a likely culprit.
	ErrCipherCreation = errors.New("could not create cipher")

	// ErrBase64DecodeError indicates a problem in base64 decoding an input
	// text.
	ErrBase64DecodeError = errors.New("could not base64 decode input")

	// ErrInvalidCipherTextBlockSize indicates a mismatch between the
	// ciphertext input length and the cipher block size.
	ErrInvalidCipherTextBlockSize = errors.New("invalid ciphertext block size")

	// ErrGeneratingIV indicates an issue during the initialisation vector
	// creation stage.
	ErrGeneratingIV = errors.New("could not create IV")
)

var supportedSymmetricAlgorithms = []string{
	"aes128",
	"aes192",
	"aes256",
	"des",
}

type Algorithm string

const (
	AES128 Algorithm = "aes128"
	AES192 Algorithm = "aes192"
	AES256 Algorithm = "aes256"
	DES    Algorithm = "des"
)

// Algorithms returns the list of supported symmetric algorithms.
func Algorithms() []string {
	return supportedSymmetricAlgorithms
}

// ValidateAlgoKeyPair takes an algorithm and a key and returns true if the key
// is valid for the given algorithm, false otherwise.
func ValidateAlgoKeyPair(algo Algorithm, key []byte) bool {
	switch algo {
	case AES128:
		return len(key) == 16
	case AES192:
		return len(key) == 24
	case AES256:
		return len(key) == 32
	case DES:
		return len(key) == 8
	default:
		return false
	}
}

// Encrypt takes an algorithm name, a key and a plaintext and attempts to
// encrypt the plaintext. If successful the base64 encoded cipher text is
// returned. Consult the typed errors in this package to understand which
// errors can occur.
//
// **CREDIT** Inspired by the code I saw here:
//
//	https://gist.github.com/fracasula/38aa1a4e7481f9cedfa78a0cdd5f1865
func Encrypt(algo Algorithm, key []byte, plaintext string) (string, error) {
	var (
		blockSize      int
		block          cipher.Block
		plaintextBytes = []byte(plaintext)
		err            error
	)
	switch algo {
	case AES128, AES192, AES256:
		blockSize = aes.BlockSize
		block, err = aes.NewCipher(key)
	case DES:
		blockSize = des.BlockSize
		block, err = des.NewCipher(key)
	}
	if err != nil {
		return "", errors.Join(ErrCipherCreation, err)
	}

	cipherText := make([]byte, blockSize+len(plaintextBytes))
	iv := cipherText[:blockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.Join(ErrGeneratingIV, err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[blockSize:], plaintextBytes)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt takes an algorithm name, a key and a cipher text and attempts to
// decrypt the cipher text. If successful the unencoded plaintext is returned.
// Consult the typed errors in this package to understand which errors can
// occur.
//
// **CREDIT** Inspired by the code I saw here:
//
//	https://gist.github.com/fracasula/38aa1a4e7481f9cedfa78a0cdd5f1865
func Decrypt(algo Algorithm, key []byte, cipherText string) (string, error) {
	var (
		blockSize int
		block     cipher.Block
		err       error
	)
	switch algo {
	case AES128, AES192, AES256:
		blockSize = aes.BlockSize
		block, err = aes.NewCipher(key)
	case DES:
		blockSize = des.BlockSize
		block, err = des.NewCipher(key)
	}
	if err != nil {
		return "", errors.Join(ErrCipherCreation, err)
	}

	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", errors.Join(ErrBase64DecodeError, err)
	}

	if len(cipherTextBytes) < blockSize {
		return "", errors.Join(ErrInvalidCipherTextBlockSize, err)
	}

	iv := cipherTextBytes[:blockSize]
	cipherTextBytes = cipherTextBytes[blockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherTextBytes, cipherTextBytes)

	return string(cipherTextBytes), nil
}
