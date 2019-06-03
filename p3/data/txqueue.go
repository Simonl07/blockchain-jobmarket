package data

/*
	This TXQueue implementation is copied from the golang api:
	https://golang.org/pkg/container/heap/

	This TXQueue is used as the TXQueue for miners to prioritize high transaction fee transactions
*/

// This example demonstrates a priority queue built using the heap interface.

import (
	"container/heap"

	"../../transaction"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value    tx.Transaction // The value of the item; arbitrary.
	Priority float32        // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	Index int // The index of the item in the heap.
}

// A TXQueue implements heap.Interface and holds Items.
type TXQueue []*Item

func (txl TXQueue) Len() int { return len(txl) }

func (txl TXQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return txl[i].Priority > txl[j].Priority
}

func (txl TXQueue) Swap(i, j int) {
	txl[i], txl[j] = txl[j], txl[i]
	txl[i].Index = i
	txl[j].Index = j
}

func (txl *TXQueue) Push(x interface{}) {
	n := len(*txl)
	item := x.(*Item)
	item.Index = n
	*txl = append(*txl, item)
}

func (txl *TXQueue) Pop() interface{} {
	old := *txl
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*txl = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (txl *TXQueue) update(item *Item, value tx.Transaction, priority float32) {
	item.Value = value
	item.Priority = priority
	heap.Fix(txl, item.Index)
}

func (txl *TXQueue) Contains(t tx.Transaction) bool {
	for _, k := range *txl {
		if k.Value.Hash == t.Hash {
			return true
		}
	}
	return false
}
