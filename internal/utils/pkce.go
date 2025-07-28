package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type PKCEChallenge struct {
	CodeVerifier  string
	CodeChallenge string
	Method        string
}

func GenerateCodeVerifier() string {
	verifier := make([]byte, 32)
	rand.Read(verifier)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(verifier)
}

func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}

func NewPKCEChallenge() *PKCEChallenge {
	verifier := GenerateCodeVerifier()
	challenge := GenerateCodeChallenge(verifier)

	return &PKCEChallenge{
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
		Method:        "S256",
	}
}
