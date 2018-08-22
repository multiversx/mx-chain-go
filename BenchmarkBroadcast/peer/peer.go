package peer

import (
	"github.com/satori/go.uuid"
)

var ct = 0

type Peer struct {
	Nr          int
	Id          uuid.UUID
	PeerMap     map[uuid.UUID]int
	Path        []int
	PathLatency int
	Latency     int
}

func New() Peer {

	var path []int

	p := Peer{
		ct,
		uuid.Must(uuid.NewV4()),
		make(map[uuid.UUID]int),
		path,
		0,
		-1,
	}
	ct++
	return p
}

func (peer Peer) SetPeersToPeerMap(distances map[uuid.UUID]int) {
	for key, value := range distances {
		peer.PeerMap[key] = value
	}
}

func (peer *Peer) SetPeerPath(path []int, senderId int) {
	for _, value := range path {
		peer.Path = append(peer.Path, value)
	}
	peer.Path = append(peer.Path, senderId)

}
func ResetCounter() {
	ct = 0
}
