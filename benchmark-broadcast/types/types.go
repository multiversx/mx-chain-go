package types

import "github.com/satori/go.uuid"

// Peer holds all details which are need for a peer
type Peer struct {
	Nr          int
	Id          uuid.UUID
	PeerMap     []uuid.UUID
	Path        []int
	PathLatency float64
	Latency     float64
}
