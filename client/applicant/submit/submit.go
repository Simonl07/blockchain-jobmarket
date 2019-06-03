package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
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
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <private key pem> <knownhost> <application json>")
		return
	}
	privatekeypem := os.Args[1]
	host := "http://" + os.Args[2]
	applicationfile := os.Args[3]

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

	fullApplication := new(models.TimestampedApplication)
	fullApplication.Application = *application
	fullApplication.Timestamp = time.Now().UnixNano() / 1000000

	fullApplicationBytes, err := json.Marshal(fullApplication)
	fullApplicationHash := sha256.Sum256(fullApplicationBytes)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, fullApplicationHash[:])
	if err != nil {
		panic(err)
	}

	signature := &models.ECDSASignature{R: r, S: s}
	signedMerit := &models.SignedMerit{Merit: application.Merit, Timestamp: fullApplication.Timestamp, Hash: hex.EncodeToString(fullApplicationHash[:]), Signature: *signature}
	signedMeritBytes, _ := json.Marshal(signedMerit)

	t := new(tx.Transaction)
	t.From = tx.EncodeECDSAPublicKey(&privateKey.PublicKey)
	t.To = ""
	t.TXType = "application"
	t.TXFee = 0.1
	t.Payload = string(signedMeritBytes)
	t.Timestamp = time.Now().UnixNano() / 1000000
	t.Hash = t.GenHash()
	t.Sign(privateKey)
	tbytes, _ := json.Marshal(t)
	http.Post(host+"/transaction", "application/json", bytes.NewBuffer(tbytes))
}
