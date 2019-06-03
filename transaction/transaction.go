package tx

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"../models"
)

// preSignedTransaction represents a Transaction without signature, used for signature verification process
type preSignedTransaction struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	TXType    string  `json:"txtype"`
	TXFee     float32 `json:"txfee"`
	Timestamp int64   `json:"timestamp"`
	Payload   string  `json:"payload"`
}

// Transaction encapsulate all the data of a transaction
type Transaction struct {
	From      string                `json:"from"`
	To        string                `json:"to"`
	TXType    string                `json:"txtype"`
	TXFee     float32               `json:"txfee"`
	Timestamp int64                 `json:"timestamp"`
	Payload   string                `json:"payload"`
	Hash      string                `json:"hash"`
	Signature models.ECDSASignature `json:"signature"`
}

// Initial initialize a new transaction with given data
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) {
	hashbytes, err := hex.DecodeString(tx.Hash)
	if err != nil {
		panic(err)
	}
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashbytes)
	if err != nil {
		panic(err)
	}
	tx.Signature = models.ECDSASignature{R: r, S: s}
}

// EncodeToJSON encode the given tx struct to string
func (tx *Transaction) EncodeToJSON() (jsonString string, error error) {
	bytes, error := json.Marshal(tx)
	jsonString = string(bytes)
	return
}

// DecodeFromJSON decode the information of transaction from jsonString into tx struct
func (tx *Transaction) DecodeFromJSON(jsonString string) {
	json.Unmarshal([]byte(jsonString), tx)
}

// GenHash generates the hash of the tx
func (tx *Transaction) GenHash() string {
	pstx := new(preSignedTransaction)
	pstx.From = tx.From
	pstx.To = tx.To
	pstx.TXType = tx.TXType
	pstx.TXFee = tx.TXFee
	pstx.Timestamp = tx.Timestamp
	pstx.Payload = tx.Payload
	bytes, _ := json.Marshal(pstx)
	pstxstr := string(bytes)
	hash := sha256.Sum256([]byte(pstxstr))
	return hex.EncodeToString(hash[:])
}

// Verify will return true if this is a valid transaction in terms of format
func (tx *Transaction) Verify() bool {
	if !tx.verifyHash() {
		return false
	}
	if !tx.verifySignature(DecodeECDSAPublicKey(tx.From)) {
		return false
	}
	return true
}

func (tx *Transaction) verifyHash() bool {
	return tx.Hash == tx.GenHash()
}

func (tx *Transaction) verifySignature(senderPubKey *ecdsa.PublicKey) bool {
	hashbytes, err := hex.DecodeString(tx.Hash)
	if err != nil {
		return false
	}
	return ecdsa.Verify(senderPubKey, hashbytes, tx.Signature.R, tx.Signature.S)
}
