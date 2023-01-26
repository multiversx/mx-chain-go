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
