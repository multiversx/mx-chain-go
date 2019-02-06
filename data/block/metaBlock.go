package block

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

// Two types of blocks
// 1. when account registers or de-registers needs to be notarized
// 2. when shards sends added block

// - Peer registration will add a node into a waiting list at epoch e
//    then it will be added to a specific shard waiting list at epoch e+1
//      at epoch e+2 it is an eligible validator

// - blocks with headers and proofs will be received from every shard each round
//   - all of this blocks will be aggregated into one metachain block

// PeerAction type represents the possible events that a node can trigger for the metachain to notarize
type PeerAction int

// Constants mapping the actions that a node can take
const (
	PeerRegistrantion PeerAction = iota + 1
	PeerDeregistration
)

func (pa PeerAction) String() string {
	switch pa {
	case PeerRegistrantion:
		return "PeerRegistration"
	case PeerDeregistration:
		return "PeerDeregistration"
	default:
		return fmt.Sprintf("Unknown type (%d)", pa)
	}
}

type peerInfo struct {
	PublicKey crypto.PublicKey
	Action PeerAction
	TimeStamp uint64
	Value *big.Int
}

type MetaBlock struct {
	// In one round it will have:
	// Shard data - should be a map? of shard/id and an array of hashes
}
