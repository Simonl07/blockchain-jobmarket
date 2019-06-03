package data

import (
	"sync"
	"time"

	"../../p1"
	"../../p2"
	"../../transaction"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	blockChain := new(p2.BlockChain)
	blockChain.Initial()
	return SyncBlockChain{bc: *blockChain}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	blocks := sbc.bc.Get(height)
	return blocks, blocks != nil
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	blocks := sbc.bc.Get(height)
	output := new(p2.Block)
	defer sbc.mux.Unlock()
	if blocks == nil {
		return *output, false
	}
	for _, v := range blocks {
		if v.Header.Hash == hash {
			return v, true
		}
	}
	return *output, false
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	blocks := sbc.bc.Get(insertBlock.Header.Height - 1)
	parentHash := insertBlock.Header.ParentHash
	for _, b := range blocks {
		if b.Header.Hash == parentHash {
			return true
		}
	}
	return false
}

func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.bc.DecodeFromJSON(blockChainJson)
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	return sbc.bc.EncodeToJSON()
}

func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie, nonce string, producer string) p2.Block {
	block := new(p2.Block)
	if sbc.bc.Length == 0 {
		block.Initial(1, time.Now().UnixNano()/1000000, "Genesis", producer, mpt)
	} else {
		lastBlock := sbc.bc.Chain[sbc.bc.Length][0]
		block.Initial(lastBlock.Header.Height+1, time.Now().UnixNano()/1000000, lastBlock.Header.Hash, producer, mpt)
	}
	block.Header.Nonce = nonce
	sbc.Insert(*block)
	return *block
}

// GetLatestBlocks returns the lastest block in bc
func (sbc *SyncBlockChain) GetLatestBlocks() ([]p2.Block, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

// GetParentBlock returns the parent block of the given block
func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) (p2.Block, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}

func (sbc *SyncBlockChain) Len() int32 {
	return sbc.bc.Len()
}

func (sbc *SyncBlockChain) ContainsBlock(block p2.Block) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.ContainsBlock(block)
}

func (sbc *SyncBlockChain) ContainsTransaction(t tx.Transaction) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.ContainsTransaction(t)
}

func (sbc *SyncBlockChain) CanonicalFromBlock(b p2.Block) []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.CanonicalFromBlock(b)
}

func (sbc *SyncBlockChain) Canonical(lookback int) ([]p2.Block, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Canonical(lookback)
}

func (sbc *SyncBlockChain) Canonicals() ([][]p2.Block, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Canonicals()
}

func (sbc *SyncBlockChain) Transactions() []tx.Transaction {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Transactions()
}

func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}
