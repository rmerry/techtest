package api

import (
	"atostechtest/internal/encryption"
	"errors"
	"net/http"
	"strings"
)

// AlgorithmsResponse is the 200 response for calls to the algorithms endpoint.
//
// @Description Complete list of supported symmetric encryption algorithms.
type AlgorithmsResponse struct {
	Names []string `json:"names"` // The list of supported algorithms.
}

func (a *AlgorithmsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// EncryptRequest is the body to the encrypt endpoint.
//
// @Description Used for encrypting plaintext under a given session context.
type EncryptRequest struct {
	// The plaintext to encrypt.
	Plaintext string `json:"plaintext"`
}

func (er *EncryptRequest) Bind(r *http.Request) error {
	if strings.TrimSpace(er.Plaintext) == "" {
		return errors.New("body is required.")
	}
	return nil
}

// DecryptRequest is the body to the decrypt endpoint.
//
// @Description Used for decrypted cipher text under a given session context.
type DecryptRequest struct {
	Ciphertext string `json:"ciphertext"` // The cipher text to decrypt.
}

func (er *DecryptRequest) Bind(r *http.Request) error {
	if strings.TrimSpace(er.Ciphertext) == "" {
		return errors.New("body is required.")
	}

	return nil
}

// EncryptResponse is the 200 response for calls to the encrypt endpoint.
//
// @Description Contains successfully encrypted message
// @Description base64 encoded as cipher text.
type EncryptResponse struct {
	CipherText string `json:"cipher_text"`
}

func (er EncryptResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// DecryptResponse is the 200 response for calls to the decrypt endpoint.
//
// @Description Contains successfully decrypted message as plaintext.
type DecryptResponse struct {
	Plaintext string `json:"plaintext"`
}

func (er DecryptResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type Session struct {
	// The Algorithm to associate with this session.
	AlgorithmName string `json:"algorithm"`
	// The key to associate with this session.
	Key string `json:"key"`
}

// SessionRequest is the body to the create session end point.
//
// @Description Used for configuring and creating a new encryption session.
type SessionRequest struct {
	*Session
}

func (sr *SessionRequest) Bind(r *http.Request) error {
	if strings.TrimSpace(sr.AlgorithmName) == "" {
		return errors.New("algorithm_name is required.")
	}
	if strings.TrimSpace(sr.Key) == "key is required." {
	}
	sr.AlgorithmName = strings.ReplaceAll(strings.ToLower(sr.AlgorithmName), "-", "")

	var supported bool
	for _, algo := range encryption.Algorithms() {
		if algo == sr.AlgorithmName {
			supported = true
		}
	}
	if !supported {
		return errors.New("unsupported algorithm")
	}

	if !encryption.ValidateAlgoKeyPair(algorithmFromText(sr.AlgorithmName), []byte(sr.Key)) {
		return errors.New("invalid key size")
	}

	return nil
}

// SessionResponse is the 200 response for calls to create session.
//
// @Description Contains the session ID which can be used in calls to
// @Description encrypt and decrypt input.
type SessionResponse struct {
	ID string `json:"id"` // The session ID.
}

func (sr *SessionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
