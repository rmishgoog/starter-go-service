package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/open-policy-agent/opa/rego"
)

var (
	//go:embed rego/authentication.rego
	opaAuthentication string
)

func main() {
	err := gentoken()
	if err != nil {
		log.Fatalln(err)
	}
}

func gentoken() error {

	privateKey, err := genkey()
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
	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)

	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	str, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println("********Signed JWT Token********")
	fmt.Println(str)
	fmt.Println("****************")

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}

	var claims2 struct {
		jwt.RegisteredClaims
		Roles []string
	}

	tkn, err := parser.ParseWithClaims(str, &claims2, keyFunc)
	if err != nil {
		return fmt.Errorf("parsing token: %w", err)
	}

	if !tkn.Valid {
		return errors.New("signature failed")
	}

	fmt.Println("SIGNATURE VALIDATED")
	fmt.Printf("%#v\n", claims2)
	fmt.Println("****************")

	var claims3 struct {
		jwt.RegisteredClaims
		Roles []string
	}

	_, _, err = parser.ParseUnverified(str, &claims3)
	if err != nil {
		return fmt.Errorf("error parsing token unver: %w", err)
	}

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	var b bytes.Buffer
	// Write the public key to the public key file.
	if err := pem.Encode(&b, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	input := map[string]any{
		"Key":   b.String(),
		"Token": str,
	}

	if err := opaPolicyEvaluation(context.Background(), opaAuthentication, input); err != nil {
		return fmt.Errorf("authentication failed : %w", err)
	}

	fmt.Println("SIGNATURE VALIDATED BY REGO")
	fmt.Println("****************")

	return nil
}

func opaPolicyEvaluation(ctx context.Context, opaPolicy string, input any) error {
	const opaPackage = "ardan.rego"
	const rule string = "auth"

	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaPolicy),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}

func genkey() (*rsa.PrivateKey, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}
	//Create a file for the private key information in PEM form.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return nil, fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	//Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	//Write the private key to the private key file.

	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return nil, fmt.Errorf("encoding to private file: %w", err)
	}

	//Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return nil, fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	//Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshaling public key: %w", err)
	}

	//Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}
	//Write the public key to the public key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return nil, fmt.Errorf("encoding to public file: %w", err)
	}
	fmt.Println("**************************************")
	fmt.Println("private and public key files generated")
	return privateKey, nil
}
