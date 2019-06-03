package data

import "encoding/json"

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {
	rd := new(RegisterData)
	rd.AssignedId = id
	rd.PeerMapJson = peerMapJson
	return *rd
}

func (data *RegisterData) EncodeToJson() (string, error) {
	bytes, err := json.Marshal(data)
	return string(bytes), err
}
