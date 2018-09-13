package p2p

import "encoding/json"

type Message struct {
	Payload string

	Hops  int
	Peers []string
}

func NewMessage(peerID string, payload string) *Message {
	return &Message{Payload: payload, Hops: 0, Peers: []string{peerID}}
}

func (m *Message) ToJson() string {
	b, err := json.Marshal(m)

	if err != nil {
		panic(err)
	}

	return string(b)
}

func FromJson(s string) *Message {
	m := Message{}
	json.Unmarshal([]byte(s), &m)

	return &m
}

func (m *Message) AddHop(peerID string) {
	m.Hops++
	m.Peers = append(m.Peers, peerID)
}
