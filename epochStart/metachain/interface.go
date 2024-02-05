package metachain

import (
	"github.com/multiversx/mx-chain-go/state"
)

// AuctionListDisplayHandler should be able to display auction list data during selection process
type AuctionListDisplayHandler interface {
	DisplayOwnersData(ownersData map[string]*OwnerAuctionData)
	DisplayOwnersSelectedNodes(ownersData map[string]*OwnerAuctionData)
	DisplayAuctionList(
		auctionList []state.ValidatorInfoHandler,
		ownersData map[string]*OwnerAuctionData,
		numOfSelectedNodes uint32,
	)
	IsInterfaceNil() bool
}
