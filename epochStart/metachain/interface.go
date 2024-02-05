package metachain

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/state"
)

type AuctionListDisplayHandler interface {
	DisplayMinRequiredTopUp(topUp *big.Int, startTopUp *big.Int)
	DisplayOwnersData(ownersData map[string]*OwnerAuctionData)
	DisplayOwnersSelectedNodes(ownersData map[string]*OwnerAuctionData)
	DisplayAuctionList(
		auctionList []state.ValidatorInfoHandler,
		ownersData map[string]*OwnerAuctionData,
		numOfSelectedNodes uint32,
	)
	IsInterfaceNil() bool
}
