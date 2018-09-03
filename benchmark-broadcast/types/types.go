package types

import "github.com/satori/go.uuid"

type Peer struct {
	Nr          int
	Id          uuid.UUID
	PeerMap     []uuid.UUID
	Path        []int
	PathLatency float64
	Latency     float64
}
