package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	privateKey, publicKey, err := generateKeyPair(4096)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	privateFile, err := os.OpenFile("id_rsa", os.O_WRONLY|os.O_CREATE, 0600)
	err = writePrivateKey(privateFile, privateKey)
	if err != nil {
		log.Fatal().Err(err).Send()
		return
	}
	err = privateFile.Close()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	publicFile, err := os.OpenFile("id_rsa.pub", os.O_WRONLY|os.O_CREATE, 0644)
	err = writePublicKey(publicFile, publicKey)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	err = publicFile.Close()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func generateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

func writePrivateKey(file *os.File, key *rsa.PrivateKey) error {
	err := pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return err
	}

	return nil
}

func writePublicKey(file *os.File, key *rsa.PublicKey) error {
	err := pem.Encode(file, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})
	if err != nil {
		return err
	}

	return nil
}
