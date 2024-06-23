package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func main() {
	err := gentoken()
	if err != nil {
		log.Fatalln(err)
	}
}

func gentoken() error {

	pk, err := genkey()
	if err != nil {
		return fmt.Errorf("error getting the private key for signing the token: %w", err)
	}

	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "12345678789",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"ADMIN"},
	}
	method := jwt.GetSigningMethod(jwt.SigningMethodES256.Name)
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	str, err := token.SignedString(pk)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println("****************")
	fmt.Println(str)
	fmt.Println("****************")

	return nil
}

func genkey() (*rsa.PrivateKey, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating private key failed with error: %w", err)
	}

	privateFile, err := os.Create("private.pem")
	if err != nil {
		return nil, fmt.Errorf("could not create the private key file: %w", err)
	}
	defer privateFile.Close()

	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return nil, fmt.Errorf("encoding private key file failed with error: %w", err)
	}

	publicFile, err := os.Create("public.pem")
	if err != nil {
		return nil, fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return nil, fmt.Errorf("encoding to public file: %w", err)
	}

	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return nil, fmt.Errorf("encoding to public file: %w", err)
	}

	return privateKey, nil
}
