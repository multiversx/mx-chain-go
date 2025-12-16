package dtos

import "github.com/multiversx/mx-chain-core-go/data"

// BroadcastData holds data to be broadcasted
type BroadcastData struct {
	Header    data.HeaderHandler
	LeaderKey []byte
	Proof     data.HeaderProofHandler
}
