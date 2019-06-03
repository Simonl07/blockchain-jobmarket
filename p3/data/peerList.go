package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

func NewPeerList(id int32, maxLength int32) PeerList {
	peerList := new(PeerList)
	peerList.selfId = id
	peerList.maxLength = maxLength
	peerList.peerMap = make(map[string]int32)
	return *peerList
}

func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	peers.peerMap[addr] = id
	peers.mux.Unlock()
}

func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (peers *PeerList) Rebalance() {
	selfID := peers.GetSelfId()
	peers.mux.Lock()
	defer peers.mux.Unlock()
	if int32(len(peers.peerMap)) <= peers.maxLength {
		return
	}

	invertedPeerMap := make(map[int32]string)
	ids := []int32{}
	for k, v := range peers.peerMap {
		invertedPeerMap[v] = k
		ids = append(ids, v)
	}
	peers.peerMap = make(map[string]int32)
	ids = append(ids, selfID)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	//find the center of the ring
	selfIndex := 0
	for i, v := range ids {
		if v == selfID {
			selfIndex = i
			break
		}
	}

	newIds := []int32{}
	idsl := len(ids)
	var left, right int
	//record upto maxLength / 2 elements from both sides
	for diff := 1; diff <= min(idsl, int(peers.maxLength/2)); diff++ {
		left = (selfIndex - diff + idsl) % idsl
		right = (selfIndex + diff) % idsl
		newIds = append(newIds, ids[left])
		newIds = append(newIds, ids[right])
	}
	for _, v := range newIds {
		peers.peerMap[invertedPeerMap[v]] = v
	}

}

func (peers *PeerList) Show() string {
	peers.mux.Lock()
	output, _ := json.Marshal(peers.peerMap)
	peers.mux.Unlock()
	return string(output)
}

func (peers *PeerList) Register(id int32) {
	peers.mux.Lock()
	peers.selfId = id
	peers.mux.Unlock()
}

func (peers *PeerList) Copy() map[string]int32 {
	peers.mux.Lock()
	tempMap := make(map[string]int32)
	for k, v := range peers.peerMap {
		tempMap[k] = v
	}
	peers.mux.Unlock()
	return tempMap
}

func (peers *PeerList) GetSelfId() int32 {
	return peers.selfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {
	return peers.Show(), nil
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	tempMap := make(map[string]int32)
	json.Unmarshal([]byte(peerMapJsonStr), &tempMap)
	peers.mux.Lock()
	for k, v := range tempMap {
		if k != selfAddr {
			peers.peerMap[k] = v
		}
	}
	peers.mux.Unlock()
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
