package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run keygen.go <output pem file> <type>")
		return
	}
	outputfile := os.Args[1]
	keytype := os.Args[2]

	var x509Encoded []byte
	if keytype == "ecdsa" {
		privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		x509Encoded, _ = x509.MarshalECPrivateKey(privateKey)
	} else if keytype == "rsa" {
		privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		x509Encoded = x509.MarshalPKCS1PrivateKey(privateKey)
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	f, err := os.Create(outputfile)
	if err != nil {
		panic(err)
	}

	l, err := f.WriteString(string(pemEncoded))
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	fmt.Println(l, "bytes written successfully to "+outputfile)
	f.Close()
}
