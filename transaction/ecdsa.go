package tx

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

/*
	Functions for serializing and deserializing ecdsa Public key to String
	this is a modified version from the original author mikijov:
	https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
*/
func EncodeECDSAPublicKey(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	pubkeystr := string(pemEncodedPub)
	return pubkeystr[27 : len(pubkeystr)-26]
}

func DecodeECDSAPublicKey(pemEncodedPub string) *ecdsa.PublicKey {
	pemEncodedPub = "-----BEGIN PUBLIC KEY-----\n" + pemEncodedPub + "\n-----END PUBLIC KEY-----\n"
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey
}

func EncodeRSAPublicKey(publicKey *rsa.PublicKey) string {
	x509EncodedPub := x509.MarshalPKCS1PublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	pubkeystr := string(pemEncodedPub)
	return pubkeystr[27 : len(pubkeystr)-26]
}

func DecodeRSAPublicKey(pemEncodedPub string) *rsa.PublicKey {
	pemEncodedPub = "-----BEGIN PUBLIC KEY-----\n" + pemEncodedPub + "\n-----END PUBLIC KEY-----\n"
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes

	rsapublicKey, err := x509.ParsePKCS1PublicKey(x509EncodedPub)
	if err != nil {
		panic(err)
	}
	return rsapublicKey
}
