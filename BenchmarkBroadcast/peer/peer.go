package peer

import (
	//"fmt"
	"github.com/satori/go.uuid"
)

var ct = 0

type Peer struct {
	Nr      int
	Id      uuid.UUID
	PeerMap map[uuid.UUID]int
}

func New() Peer {

	p := Peer{
		ct,
		uuid.Must(uuid.NewV4()),
		make(map[uuid.UUID]int),
	}
	ct++
	return p
}

func (peer Peer) SetPeersToPeerMap(distances map[uuid.UUID]int) {
	for key, value := range distances {
		peer.PeerMap[key] = value
	}
}
