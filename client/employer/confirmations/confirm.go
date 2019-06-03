package main

import (
	"crypto/ecdsa"
	"crypto/rand"
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

	"../../../models"
	"../../../transaction"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <rsa pem> <knownhost> <merithash>")
		return
	}
	rsapem := os.Args[1]
	host := "http://" + os.Args[2]
	merithash := os.Args[3]

	rsaprivatekeybytes, _ := ioutil.ReadFile(rsapem)
	rsablock, _ := pem.Decode(rsaprivatekeybytes)
	rsax509Encoded := rsablock.Bytes
	rsapk, _ := x509.ParsePKCS1PrivateKey(rsax509Encoded)

	resp, err := http.Get(host + "/transactions")
	if err != nil {
		fmt.Println("Cannot connect to known host")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	transactions := make([]tx.Transaction, 0)
	json.Unmarshal(body, &transactions)

	var acceptance tx.Transaction
	var meritTransaction tx.Transaction
	found := false
	sm := new(models.SignedMerit)
	for _, t := range transactions {
		if t.TXType == "application" {
			json.Unmarshal([]byte(t.Payload), &sm)
			if sm.Hash == merithash {
				meritTransaction = t
			}
		}
		if t.To == merithash && !found {
			acceptance = t
			found = true
		}
	}
	if found {
		fmt.Println("Your acceptance hash: " + acceptance.Hash)
		for _, t := range transactions {
			if t.To == acceptance.Hash {
				fmt.Printf("Merit %v has been accepted by you and confirmed by applicant\n", merithash)
				ciphertext, err := hex.DecodeString(t.Payload)
				if err != nil {
					panic(err)
				}
				decryptedIdentityBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsapk, ciphertext, []byte(""))
				if err != nil {
					panic(err)
				}
				fmt.Printf("decrypted identity: %v\n", string(decryptedIdentityBytes))
				fmt.Println("Assembling original application for verification...")

				identity := new(models.Identity)
				json.Unmarshal(decryptedIdentityBytes, &identity)

				signedMerit := new(models.SignedMerit)
				json.Unmarshal([]byte(meritTransaction.Payload), &signedMerit)

				application := new(models.Application)
				application.Merit = signedMerit.Merit
				application.Identity = *identity

				fullApplication := new(models.TimestampedApplication)
				fullApplication.Application = *application
				fullApplication.Timestamp = signedMerit.Timestamp

				fullApplicationBytes, _ := json.Marshal(fullApplication)
				fullApplicationBytesDisplay, _ := json.MarshalIndent(fullApplication, "", "\t")
				fullApplicationHash := sha256.Sum256(fullApplicationBytes)

				applicantPubKey := tx.DecodeECDSAPublicKey(meritTransaction.From)
				valid := ecdsa.Verify(applicantPubKey, fullApplicationHash[:], signedMerit.Signature.R, signedMerit.Signature.S)
				if valid {
					fmt.Printf("Signature verified, decrypted identity is authentic, full application: \n%v\n", string(fullApplicationBytesDisplay))
				} else {
					fmt.Println("Signature verification failed, Illegal confirmation due to inauthentic indentity claim. Disregard ")
				}
				return
			}
		}
		fmt.Println("The applicant has not yet confirmed your acceptance")
	} else {
		fmt.Printf("You have not accepted merit %v or merits does not exist\n", merithash)
	}

}
