package p2

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"../p1"
	"../transaction"
	"golang.org/x/crypto/sha3"
)

type Block struct {
	Header BlockHeader
	Value  p1.MerklePatriciaTrie `json:"mpt"`
}

type BlockHeader struct {
	Nonce      string
	Height     int32
	Timestamp  int64
	Hash       string
	ParentHash string
	Size       int32
	Producer   string
}

type BlockJsonFormat struct {
	Nonce      string            `json:"nonce"`
	Height     int32             `json:"height"`
	Timestamp  int64             `json:"timestamp"`
	Hash       string            `json:"hash"`
	ParentHash string            `json:"parentHash"`
	Size       int32             `json:"size"`
	Producer   string            `json:"producer"`
	Value      map[string]string `json:"value"`
}

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func (block *Block) Initial(height int32, timestamp int64, parentHash string, producer string, value p1.MerklePatriciaTrie) {
	header := new(BlockHeader)
	header.Height = height
	header.Timestamp = timestamp
	header.ParentHash = parentHash
	header.Producer = producer
	bytes, _ := json.Marshal(value)
	header.Size = int32(len(bytes))
	block.Header = *header
	block.Value = value
	hashStr := strconv.Itoa(int(block.Header.Height)) + strconv.Itoa(int(block.Header.Timestamp)) + block.Header.ParentHash + block.Value.Root + strconv.Itoa(int(block.Header.Size))
	sum := sha3.Sum256([]byte(hashStr))
	block.Header.Hash = hex.EncodeToString(sum[:])
}

func (block *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(&BlockJsonFormat{
		Height:     block.Header.Height,
		Timestamp:  block.Header.Timestamp,
		Hash:       block.Header.Hash,
		ParentHash: block.Header.ParentHash,
		Size:       block.Header.Size,
		Value:      block.Value.Mapping,
		Nonce:      block.Header.Nonce,
		Producer:   block.Header.Producer,
	})
}

func (block *Block) UnmarshalJSON(bytes []byte) error {
	blockJson := new(BlockJsonFormat)
	json.Unmarshal(bytes, &blockJson)
	block.Header.Height = blockJson.Height
	block.Header.ParentHash = blockJson.ParentHash
	block.Header.Timestamp = blockJson.Timestamp
	block.Header.Hash = blockJson.Hash
	block.Header.Size = blockJson.Size
	block.Header.Nonce = blockJson.Nonce
	block.Header.Producer = blockJson.Producer
	mpt := new(p1.MerklePatriciaTrie)
	mpt.Initial()
	for k, v := range blockJson.Value {
		mpt.Insert(k, v)
	}
	block.Value = *mpt
	return nil
}

func (block *Block) DecodeFromJson(jsonString string) {
	json.Unmarshal([]byte(jsonString), block)
}

func (block *Block) EncodeToJSON() (jsonString string) {
	bytes, _ := json.Marshal(block)
	jsonString = string(bytes)
	return
}

func (bc *BlockChain) Initial() {
	bc.Chain = make(map[int32][]Block)
	bc.Length = 0
}

func (bc *BlockChain) Insert(block Block) {
	blocks := bc.Chain[block.Header.Height]

	for _, b := range blocks {
		if b.Header.Hash == block.Header.Hash {
			return
		}
	}
	bc.Chain[block.Header.Height] = append(blocks, block)
	//Update block chain length if necessary
	if block.Header.Height > bc.Length {
		bc.Length = block.Header.Height
	}
}

func (bc *BlockChain) Get(height int32) []Block {
	if height > int32(len(bc.Chain)) || height < 0 {
		return nil
	}
	return bc.Chain[height]
}

func (bc *BlockChain) Len() int32 {
	return bc.Length
}

func (bc *BlockChain) EncodeToJSON() (jsonString string, error error) {
	bytes, error := json.Marshal(bc)
	jsonString = string(bytes)
	return
}

func (bc *BlockChain) DecodeFromJSON(jsonString string) {
	blocks := []Block{}
	json.Unmarshal([]byte(jsonString), &blocks)
	for _, b := range blocks {
		bc.Insert(b)
	}
}

func (bc *BlockChain) MarshalJSON() ([]byte, error) {
	blockChainJson := []Block{}
	for i := int32(1); i <= bc.Length; i++ {
		for _, b := range bc.Chain[i] {
			blockChainJson = append(blockChainJson, b)

		}
	}
	return json.Marshal(blockChainJson)
}
func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

// GetLatestBlocks returns the lastest block in bc
func (bc *BlockChain) GetLatestBlocks() ([]Block, error) {
	blocks := bc.Get(bc.Length)
	if blocks == nil {
		return blocks, errors.New("empty block chain")
	}
	return blocks, nil
}

// GetParentBlock returns the parent block of the given block
func (bc *BlockChain) GetParentBlock(block Block) (Block, error) {
	parentHash := block.Header.ParentHash
	parentHeight := block.Header.Height - 1
	blocks := bc.Get(parentHeight)
	if blocks == nil {
		return Block{}, errors.New("no parent blocks")
	}
	for _, b := range blocks {
		if b.Header.Hash == parentHash {
			return b, nil
		}
	}
	return Block{}, errors.New("parent block does not exist")
}

/*
	What is a canonical chain at 'height' = 15 and 'lookback' = 6:

	***
		THIS VERSION OF CANONICAL CHAIN IS ONLY FOR VIEWING PURPOSE,
		NOT FOR VERIFYING BLOCKS AND TRANSACTION PURPOSE
		USE Canonicals() instead
	***

	1. That height 15 does not contains forks, if so, return C(14, 6)
	2. Once found the height with no forks, step back lookback amount of blocks
		to ensure finality
*/
func (bc *BlockChain) Canonical(lookback int) ([]Block, error) {
	return bc.canonical(bc.Length, lookback, "")
}

//Canonicals returns the canonical chains in 2D slice,
//parallel cononical chains due to forks are included
func (bc *BlockChain) Canonicals() ([][]Block, error) {
	latestBlocks, err := bc.GetLatestBlocks()
	canonicals := make([][]Block, len(latestBlocks))
	if err != nil {
		return canonicals, err
	}
	for i, b := range latestBlocks {
		canonicals[i] = bc.CanonicalFromBlock(b)
	}
	return canonicals, nil
}

func (bc *BlockChain) canonical(height int32, lookback int, hash string) ([]Block, error) {
	latestBlocks, err := bc.GetLatestBlocks()
	if err != nil {
		return make([]Block, 0), err
	}

	if len(latestBlocks) > 1 && hash == "" {
		return bc.canonical(height-1, lookback, "")
	}

	if lookback == 0 {
		var block Block
		if hash == "" {
			block = latestBlocks[0]
		} else {
			for _, b := range latestBlocks {
				if b.Header.Hash == hash {
					block = b
				}
			}
		}
		return bc.CanonicalFromBlock(block), nil
	}
	block := latestBlocks[0]
	return bc.canonical(height-1, lookback-1, block.Header.ParentHash)
}

func (bc *BlockChain) CanonicalFromBlock(b Block) []Block {
	canonicalchain := make([]Block, 0)
	canonicalchain = append(canonicalchain, b)
	b, err := bc.GetParentBlock(b)
	for err == nil {
		canonicalchain = append(canonicalchain, b)
		b, err = bc.GetParentBlock(b)
	}
	return canonicalchain
}

func (bc *BlockChain) Transactions() []tx.Transaction {
	canonicalchain, _ := bc.Canonical(0)
	transactions := make([]tx.Transaction, 0)
	for _, b := range canonicalchain {
		for k := range b.Value.Mapping {
			v, _ := b.Value.Get(k)
			t := new(tx.Transaction)
			t.DecodeFromJSON(v)
			transactions = append(transactions, *t)
		}
	}
	return transactions
}

func (bc *BlockChain) ContainsTransaction(t tx.Transaction) bool {
	canonicals, err := bc.Canonicals()
	if err != nil {
		return false
	}
	for _, c := range canonicals {
		for _, b := range c {
			for k := range b.Value.Mapping {
				if k == t.Hash {
					return true
				}
			}
		}
	}
	return false
}

func (bc *BlockChain) ContainsBlock(block Block) bool {
	canonicals, err := bc.Canonicals()
	if err != nil {
		return false
	}
	for _, c := range canonicals {
		for _, b := range c {
			if b.Header.Height == block.Header.Height && b.Header.Hash == block.Header.Hash {
				return true
			}
		}
	}
	return false
}
