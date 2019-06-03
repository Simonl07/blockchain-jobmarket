package p3

import (
	"bytes"
	"container/heap"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../models"
	"../p1"
	"../p2"
	"../transaction"
	"./data"
	"golang.org/x/crypto/sha3"
)

var DIFFICULTY = "000000"

var FIRST_NODE_HOST string
var PORT string
var NODEID string

var bcDownloadServer string
var selfAddr string

var SBC data.SyncBlockChain
var Peers data.PeerList
var TransactionQueue data.TXQueue
var ifStarted bool
var nodeID int32

var minerPrivateKey *ecdsa.PrivateKey

func init() {
	// This function will be executed before everything else.
	// Do some initialization here.
	SBC = data.NewBlockChain()
	ifStarted = false
	minerPrivateKey, _ = ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

	go func() {
		time.Sleep(time.Duration(3) * time.Second)
		go start()
	}()
}

func start() {
	bcDownloadServer = FIRST_NODE_HOST + "/upload"
	selfAddr = "http://localhost:" + PORT
	temp, _ := strconv.ParseInt(NODEID, 0, 32)
	nodeID = int32(temp)
	Peers = data.NewPeerList(nodeID, 32)
	TransactionQueue = make(data.TXQueue, 0)
	heap.Init(&TransactionQueue)
	if FIRST_NODE_HOST == "http://" {
		go StartHeartBeat()
	} else {
		Download()
		go StartHeartBeat()
	}
	go StartTryingNonces()
	ifStarted = true
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	if ifStarted {
		fmt.Fprintln(w, "already started")
		return
	}
	start()

}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Download blockchain from TA server
func Download() {
	peerMapJSON, _ := Peers.PeerMapToJson()
	hbd := data.NewHeartBeatData(nodeID, false, "", false, "", peerMapJSON, selfAddr)
	hbdJSON, _ := json.Marshal(hbd)
	res, err := http.Post(bcDownloadServer, "application/json", bytes.NewBuffer(hbdJSON))
	if err != nil {
		log.Fatal("Cannot download from first node")
	}

	json, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	SBC.UpdateEntireBlockChain(string(json))
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	hbd := new(data.HeartBeatData)
	json.Unmarshal(body, &hbd)
	Peers.Add(hbd.Addr, hbd.Id)
	blockChainJSON, err := SBC.BlockChainToJson()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, blockChainJSON)
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	heightStr := strings.Split(r.URL.Path, "/")[2]
	height, err := strconv.ParseInt(heightStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hash := strings.Split(r.URL.Path, "/")[3]
	block, success := SBC.GetBlock(int32(height), hash)
	if success {
		fmt.Fprint(w, block.EncodeToJSON())
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	//Unmarshal json
	hbd := new(data.HeartBeatData)
	json.Unmarshal(body, &hbd)

	//Add addresses to peer list
	if hbd.Addr != selfAddr {
		Peers.Add(hbd.Addr, hbd.Id)
	}
	Peers.InjectPeerMapJson(hbd.PeerMapJson, selfAddr)

	if hbd.IfNewBlock {
		block := new(p2.Block)
		block.DecodeFromJson(hbd.BlockJson)
		if !SBC.CheckParentHash(*block) {
			AskForBlock(block.Header.Height-1, block.Header.ParentHash)
		}
		if verifyBlock(*block) {
			SBC.Insert(*block)
		}
	} else if hbd.IfNewTransaction {
		processNewTransaction(hbd.TransactionJson)
	}

	if hbd.Hops > 0 {
		ForwardHeartBeat(*hbd)
	}
	//END OF HEARTBEAT
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	pm := Peers.Copy()
	for k := range pm {
		url := k + "/block/" + string(height) + "/" + hash
		res, err := http.Get(url)
		if err != nil {
			//If encounter http erros, move on to next peer to ask for block
			continue
		}
		defer res.Body.Close()
		if res.StatusCode == 200 {
			json, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			block := new(p2.Block)
			block.DecodeFromJson(string(json))
			SBC.Insert(*block)
			if !SBC.CheckParentHash(*block) {
				AskForBlock(block.Header.Height-1, block.Header.ParentHash)
			}
			return //Return as soon as the block is found
		}
	}
}

func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	Peers.Rebalance()
	peerMapJSON, _ := Peers.PeerMapToJson()
	heartBeatData.PeerMapJson = peerMapJSON
	heartBeatData.Addr = selfAddr
	heartBeatData.Id = nodeID
	heartBeatData.Hops--
	hbdJSON, _ := json.Marshal(heartBeatData)

	pm := Peers.Copy()
	for k := range pm {
		http.Post(k+"/heartbeat/receive", "application/json", bytes.NewBuffer(hbdJSON))
	}
}

func Canonical(w http.ResponseWriter, r *http.Request) {
	latestBlocks, _ := SBC.GetLatestBlocks()

	for i, b := range latestBlocks {
		fmt.Fprintf(w, "Chain #%d: \n\n", i)
		temp := b
		height := SBC.Len()
		for true {
			fmt.Fprintf(w, "Height: %d, block: %s\n\n", height, temp.EncodeToJSON())
			height--
			var err error
			temp, err = SBC.GetParentBlock(temp)
			if err != nil {
				break
			}
		}
		fmt.Fprintf(w, "\n\n")
	}
}

func ViewTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := SBC.Transactions()
	json, _ := json.MarshalIndent(transactions, "", "\t")
	fmt.Fprintln(w, string(json))
}

func ViewMerits(w http.ResponseWriter, r *http.Request) {
	transactions := SBC.Transactions()
	merits := make([]models.SignedMerit, 0)
	for _, t := range transactions {
		if t.TXType == "application" {
			sm := new(models.SignedMerit)
			json.Unmarshal([]byte(t.Payload), &sm)
			merits = append(merits, *sm)
		}
	}
	json, _ := json.MarshalIndent(merits, "", "\t")
	fmt.Fprintln(w, string(json))
}

func MinerBalance(w http.ResponseWriter, r *http.Request) {
	canonical, _ := SBC.Canonical(0)
	balancemap := make(map[string]float32)
	for _, b := range canonical {
		producer := b.Header.Producer
		txtotal := float32(0)
		t := new(tx.Transaction)
		for _, v := range b.Value.Mapping {
			json.Unmarshal([]byte(v), &t)
			txtotal += t.TXFee
		}
		if balance, ok := balancemap[producer]; ok {
			balancemap[producer] = balance + txtotal
		} else {
			balancemap[producer] = txtotal
		}
	}
	outputbytes, _ := json.MarshalIndent(balancemap, "", "\t")
	fmt.Fprintln(w, string(outputbytes))
}

func ReceiveTransaction(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	processNewTransaction(string(body))
	hbd := new(data.HeartBeatData)
	hbd.IfNewTransaction = true
	hbd.TransactionJson = string(body)
	hbd.Hops = 2
	ForwardHeartBeat(*hbd)
}

func processNewTransaction(transactionJSON string) {
	t := new(tx.Transaction)
	t.DecodeFromJSON(string(transactionJSON))

	if TransactionQueue.Contains(*t) {
		fmt.Printf("Ignored duplicate transaction %v\n", t.Hash)
		return
	}

	if verifyTransaction(*t) {
		fmt.Printf("Received valid transaction %v\n", t.Hash)
		item := &data.Item{Value: *t, Priority: t.TXFee}
		heap.Push(&TransactionQueue, item)
	} else {
		fmt.Println("Received invalid transaction, ignored")

	}
}

func pullTransactions(size int) []tx.Transaction {
	txs := make([]tx.Transaction, 0)
	i := 0
	for ; i < size && TransactionQueue.Len() > 0; i++ {
		t := heap.Pop(&TransactionQueue).(*data.Item).Value
		if !SBC.ContainsTransaction(t) {
			txs = append(txs, t)
		} else {
			i--
		}
	}
	return txs
}

//Increment a hex string by value of one
func incrementHex(input string) string {
	d, _ := strconv.ParseInt("0x"+input, 0, 64)
	d++
	return strconv.FormatInt(d, 16)
}

//Generate a random hex string with n characters
func genHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func verifyTransaction(tx tx.Transaction) bool {
	//Tx has correct hash & signature
	if !tx.Verify() {
		return false
	}

	//Tx is not in canonicalchain
	if SBC.ContainsTransaction(tx) {
		return false
	}

	//Make sure that if this is an acceptance, accepting non-existing merits is not valid
	transactions := SBC.Transactions()
	if tx.TXType == "acceptance" {
		found := false
		sm := new(models.SignedMerit)
		for _, t := range transactions {
			json.Unmarshal([]byte(t.Payload), &sm)
			if sm.Hash == tx.To {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	//Make sure that if this is a confirmation, this is confirming on a existing acceptance
	if tx.TXType == "confimation" {
		found := false
		for _, t := range transactions {
			if t.Hash == tx.To {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func verifyBlock(block p2.Block) bool {
	sum := sha3.Sum256([]byte(block.Header.ParentHash + block.Header.Nonce + block.Value.Root))
	hash := hex.EncodeToString(sum[:])
	if !strings.HasPrefix(hash, DIFFICULTY) {
		return false
	}

	if len(block.Value.Mapping) > 20 {
		fmt.Printf("Received invalid block %v (invalid block size)\n", block.Header.Hash)
		return false
	}

	if SBC.ContainsBlock(block) {
		fmt.Printf("Received existing block %v, ignored\n", block.Header.Hash)
		return false
	}

	for k := range block.Value.Mapping {
		t := new(tx.Transaction)
		tjson, _ := block.Value.Get(k)
		t.DecodeFromJSON(tjson)
		if !verifyTransaction(*t) {
			return false
		}
	}
	fmt.Printf("Received valid block %v\n", block.Header.Hash)
	return true
}

func StartHeartBeat() {
	for true {
		time.Sleep(time.Duration(rand.Intn(5)+5) * time.Second)
		Peers.Rebalance()
		peerMapJSON, _ := Peers.PeerMapToJson()
		pm := Peers.Copy()
		hbd := data.PrepareHeartBeatData(&SBC, nodeID, peerMapJSON, selfAddr)
		hbdJSON, _ := json.Marshal(hbd)
		for k := range pm {
			http.Post(k+"/heartbeat/receive", "application/json", bytes.NewBuffer(hbdJSON))
		}
	}
}

func StartTryingNonces() {
	//Seed the rand module
	rand.Seed(time.Now().UTC().UnixNano())
	nonce := genHex(16)
	prevHeight := SBC.Len()
	txs := pullTransactions(20)
	mpt := new(p1.MerklePatriciaTrie)
	mpt.Initial()
	for _, t := range txs {
		// MPT<TransactionHash, Transaction>
		tjson, _ := t.EncodeToJSON()
		mpt.Insert(t.Hash, string(tjson))
	}
	for true {
		// If no transaction was pulled from tx list, sleep for few seconds and retry
		if len(txs) == 0 {
			fmt.Printf("Transaction Queue is empty, listening for transactions\n")
			time.Sleep(time.Duration(7) * time.Second)
			txs = pullTransactions(20)
			mpt.Initial()
			for _, t := range txs {
				// MPT<TransactionHash, Transaction>
				tjson, _ := t.EncodeToJSON()
				mpt.Insert(t.Hash, string(tjson))
			}
			if len(txs) != 0 {
				fmt.Printf("Building block with %d transactions...\n", len(txs))
			}
			continue
		}

		//New block has arrived, release txs back into transaction queue and re-pull 20 transactions and make sure they are not in chain
		if prevHeight != SBC.Len() {
			prevHeight = SBC.Len()
			for _, t := range txs {
				if !SBC.ContainsTransaction(t) {
					item := &data.Item{Value: t, Priority: t.TXFee}
					heap.Push(&TransactionQueue, item)
				}
			}
			txs = pullTransactions(20)
			if len(txs) == 0 {
				continue
			}
			mpt.Initial()
			for _, t := range txs {
				// MPT<TransactionHash, Transaction>
				tjson, _ := t.EncodeToJSON()
				mpt.Insert(t.Hash, string(tjson))
			}
			fmt.Printf("Building block with %d transactions...\n", len(txs))
		}
		//Now its safe to start testing this nonce
		parentBlocks, err := SBC.GetLatestBlocks()
		var parentBlockHash string
		if err != nil {
			parentBlockHash = "Genesis"
		} else {
			parentBlock := parentBlocks[0]
			parentBlockHash = parentBlock.Header.Hash
		}
		hashStr := parentBlockHash + nonce + mpt.Root
		sum := sha3.Sum256([]byte(hashStr))
		hash := hex.EncodeToString(sum[:])
		if strings.HasPrefix(hash, DIFFICULTY) {
			fmt.Println("FOUND NONCE: " + nonce + " Hash: " + hash)
			block := SBC.GenBlock(*mpt, nonce, tx.EncodeECDSAPublicKey(&minerPrivateKey.PublicKey))
			fmt.Println("Generated block " + block.Header.Hash)
			hbd := new(data.HeartBeatData)
			hbd.IfNewBlock = true
			hbd.BlockJson = block.EncodeToJSON()
			hbd.Hops = 2
			ForwardHeartBeat(*hbd)
			txs = pullTransactions(20)
			mpt.Initial()
			for _, t := range txs {
				// MPT<TransactionHash, Transaction>
				tjson, _ := t.EncodeToJSON()
				mpt.Insert(t.Hash, string(tjson))
			}
		}
		nonce = genHex(16)
	}
}
