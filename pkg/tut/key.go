package tut

import (
	"crypto/rsa"
	"io/ioutil"
	"os"

	jwtgo "github.com/golang-jwt/jwt/v4"
)

// ReadPrivateKey reads a private key from the given file path.
func ReadPrivateKey(privateKeyPath string) *rsa.PrivateKey {
	file, err := os.Open(privateKeyPath)
	if err != nil {
		panic(err)
	}

	defer func() { _ = file.Close() }()

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	privateKey, err := jwtgo.ParseRSAPrivateKeyFromPEM(fileContent)
	if err != nil {
		panic(err)
	}

	return privateKey
}
