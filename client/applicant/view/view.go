package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"../../../models"
	"../../../transaction"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <private key pem> <knownhost>")
		return
	}
	privatekeypem := os.Args[1]
	host := "http://" + os.Args[2]

	privatekeybytes, _ := ioutil.ReadFile(privatekeypem)
	block, _ := pem.Decode(privatekeybytes)
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)
	pubKeyStr := tx.EncodeECDSAPublicKey(&privateKey.PublicKey)

	resp, err := http.Get(host + "/transactions")
	if err != nil {
		fmt.Println("Cannot connect to known host")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	transactions := make([]tx.Transaction, 0)
	json.Unmarshal(body, &transactions)

	fmt.Println("Displaying information about applicant: " + strings.Replace(pubKeyStr, "\n", "\\n", -1))
	fmt.Println("In-chain merits: ")
	merits := make([]models.SignedMerit, 0)
	for _, t := range transactions {
		if t.From == pubKeyStr {
			m := new(models.SignedMerit)
			json.Unmarshal([]byte(t.Payload), &m)
			merits = append(merits, *m)
			fmt.Println("\t" + t.Payload)
		}
	}

	fmt.Println()
	fmt.Println("Merits are currently accepted by: ")
	cnt := 1
	for _, t := range transactions {
		for _, m := range merits {
			if t.To == m.Hash {
				fmt.Printf("%d. \tMerit %v \n\tis accepted by %v\n\tin transaction %v\n", cnt, m.Hash, strings.Replace(t.From, "\n", "\\n", -1), t.Hash)
				cnt++
			}
		}
	}
}
