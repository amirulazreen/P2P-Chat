package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type Encryptor interface {
	Encrypt(msg []byte, pub *rsa.PublicKey) ([]byte, error)
	Decrypt(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error)
}

type RSAEncryptor struct{}

func (e *RSAEncryptor) Encrypt(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha256.New()
	return rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
}

func (e *RSAEncryptor) Decrypt(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.New()
	return rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
}