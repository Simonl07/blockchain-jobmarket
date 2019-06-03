package data

type HeartBeatData struct {
	IfNewBlock       bool   `json:"ifNewBlock"`
	IfNewTransaction bool   `json:"ifNewTransaction"`
	Id               int32  `json:"id"`
	BlockJson        string `json:"blockJson"`
	TransactionJson  string `json:"transactionJson"`
	PeerMapJson      string `json:"peerMapJson"`
	Addr             string `json:"addr"`
	Hops             int32  `json:"hops"`
}

func NewHeartBeatData(id int32, ifNewBlock bool, blockJson string, ifNewTransaction bool, transactionJson string, peerMapJson string, addr string) HeartBeatData {
	hbd := new(HeartBeatData)
	hbd.IfNewBlock = ifNewBlock
	hbd.IfNewTransaction = ifNewTransaction
	hbd.Id = id
	hbd.BlockJson = blockJson
	hbd.TransactionJson = transactionJson
	hbd.PeerMapJson = peerMapJson
	hbd.Addr = addr
	hbd.Hops = 2
	return *hbd
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapBase64 string, addr string) HeartBeatData {
	return NewHeartBeatData(selfId, false, "", false, "", peerMapBase64, addr)
}
