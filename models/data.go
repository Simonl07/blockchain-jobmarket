package models

import (
	"math/big"
)

type Identity struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
	Email   string `json:"email"`
}
type Merit struct {
	Experience []string `json:"experience"`
	Education  []string `json:"education"`
}

//The Hash in signedMerit is using full application hash, this is only used as indentifier
type SignedMerit struct {
	Merit     Merit          `json:"merit"`
	Hash      string         `json:"hash"`
	Timestamp int64          `json:"timestamp"`
	Signature ECDSASignature `json:"application_signature"`
}

type Application struct {
	Identity Identity `json:"identity"`
	Merit    Merit    `json:"merit"`
}

type TimestampedApplication struct {
	Application Application `json:"application"`
	Timestamp   int64       `json:"timestamp"`
}

// ECDSASignature encapsulate the two big.Int that is used to represent the signature body
type ECDSASignature struct {
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
}
