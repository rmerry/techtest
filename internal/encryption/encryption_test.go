package encryption

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		algo      Algorithm
		key       string
		plaintext string
	}{
		{AES128, "0123456789abcdef", "I'll be back"},
		{AES192, "0123456789abcdefghijklmo", "Hasta la vista, baby"},
		{AES256, "0123456789abcdefghijklmopqrstuvw", "Get to the chopper!"},
		{DES, "01234567", "I need your clothes, your boots, and your motorcycle"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Algorithm: %s", tc.algo), func(t *testing.T) {
			cipherText, err := Encrypt(tc.algo, []byte(tc.key), tc.plaintext)
			if err != nil {
				t.Errorf("encryption failed: %v", err)
			}

			if !isBase64(cipherText) {
				t.Errorf("expected ciphertext to be base64 encoded, got %q", cipherText)
			}

			// Decrypt the cipher text
			decryptedText, err := Decrypt(tc.algo, []byte(tc.key), cipherText)
			if err != nil {
				t.Errorf("decryption failed: %v", err)
			}
			if decryptedText != tc.plaintext {
				t.Errorf("decrypted text does not match plaintext, expected: %q, got: %q",
					tc.plaintext,
					decryptedText)
			}
		})
	}
}

func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
