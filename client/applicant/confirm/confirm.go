package main

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"../../../models"
	"../../../transaction"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run main.go <private key pem> <knownhost> <acceptance hash> <application json> ")
		return
	}
	privatekeypem := os.Args[1]
	host := "http://" + os.Args[2]
	acceptanceHash := os.Args[3]
	applicationfile := os.Args[4]

	privatekeybytes, _ := ioutil.ReadFile(privatekeypem)
	block, _ := pem.Decode(privatekeybytes)
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	jsonFile, err := os.Open(applicationfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	application := new(models.Application)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &application)

	identitybytes, _ := json.Marshal(application.Identity)

	resp, err := http.Get(host + "/transactions")
	if err != nil {
		fmt.Println("Cannot connect to known host")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	transactions := make([]tx.Transaction, 0)
	json.Unmarshal(body, &transactions)

	var employerPubKeyStr string
	for _, t := range transactions {
		if t.Hash == acceptanceHash {
			employerPubKeyStr = t.Payload

		}
	}
	employerPubKey := tx.DecodeRSAPublicKey(employerPubKeyStr)
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), crand.Reader, employerPubKey, identitybytes, []byte(""))

	t := new(tx.Transaction)
	t.From = tx.EncodeECDSAPublicKey(&privateKey.PublicKey)
	t.To = acceptanceHash
	t.TXType = "confirmation"
	t.TXFee = 0.1
	t.Payload = hex.EncodeToString(ciphertext)
	t.Timestamp = time.Now().UnixNano() / 1000000
	t.Hash = t.GenHash()
	t.Sign(privateKey)
	tbytes, _ := json.Marshal(t)
	http.Post(host+"/transaction", "application/json", bytes.NewBuffer(tbytes))
}
