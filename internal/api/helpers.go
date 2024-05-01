package api

import "asostechtest/internal/encryption"

// algorithmFromText takes a text input (as will come via the API) algorithm
// name and turns it into a hard coded Algorithm type as understood by the
// encryption package.
func algorithmFromText(input string) encryption.Algorithm {
	switch input {
	case "aes128":
		return encryption.AES128
	case "aes192":
		return encryption.AES192
	case "aes256":
		return encryption.AES256
	default:
		return encryption.DES
	}
}
