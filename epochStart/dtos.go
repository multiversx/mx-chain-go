package epochStart

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/state"
)

// OwnerData is a struct containing relevant information about owner's nodes data
type OwnerData struct {
	NumStakedNodes int64
	NumActiveNodes int64
	TotalTopUp     *big.Int
	TopUpPerNode   *big.Int
	AuctionList    []state.ValidatorInfoHandler
	Qualified      bool
}

// ValidatorStatsInEpoch holds validator stats in an epoch
type ValidatorStatsInEpoch struct {
	Eligible map[uint32]int
	Waiting  map[uint32]int
	Leaving  map[uint32]int
}
