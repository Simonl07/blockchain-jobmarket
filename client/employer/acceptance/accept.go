package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"../../../transaction"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run accept.go <ecdsa pr_key> <rsa pr_key> <knownhost> <merithash>")
		return
	}
	ecdsapkpem := os.Args[1]
	rsapkpem := os.Args[2]
	host := "http://" + os.Args[3]
	merithash := os.Args[4]

	ecdsaprivatekeybytes, _ := ioutil.ReadFile(ecdsapkpem)
	block, _ := pem.Decode(ecdsaprivatekeybytes)
	x509Encoded := block.Bytes
	ecdsapk, _ := x509.ParseECPrivateKey(x509Encoded)

	rsaprivatekeybytes, _ := ioutil.ReadFile(rsapkpem)
	rsablock, _ := pem.Decode(rsaprivatekeybytes)
	rsax509Encoded := rsablock.Bytes
	rsapk, _ := x509.ParsePKCS1PrivateKey(rsax509Encoded)

	t := new(tx.Transaction)
	t.From = tx.EncodeECDSAPublicKey(&ecdsapk.PublicKey)
	t.To = merithash
	t.TXType = "acceptance"
	t.TXFee = 0.1
	t.Payload = tx.EncodeRSAPublicKey(&rsapk.PublicKey)
	t.Timestamp = time.Now().UnixNano() / 1000000
	t.Hash = t.GenHash()
	t.Sign(ecdsapk)
	tbytes, _ := json.Marshal(t)
	http.Post(host+"/transaction", "application/json", bytes.NewBuffer(tbytes))
}
