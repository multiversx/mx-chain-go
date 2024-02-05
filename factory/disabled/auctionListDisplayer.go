package disabled

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/state"
)

type auctionListDisplayer struct {
}

func NewDisabledAuctionListDisplayer() *auctionListDisplayer {
	return &auctionListDisplayer{}
}

func (ald *auctionListDisplayer) DisplayMinRequiredTopUp(_ *big.Int, _ *big.Int) {

}

func (ald *auctionListDisplayer) DisplayOwnersData(_ map[string]*metachain.OwnerAuctionData) {

}

func (ald *auctionListDisplayer) DisplayOwnersSelectedNodes(_ map[string]*metachain.OwnerAuctionData) {

}

func (ald *auctionListDisplayer) DisplayAuctionList(
	_ []state.ValidatorInfoHandler,
	_ map[string]*metachain.OwnerAuctionData,
	_ uint32,
) {
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ald *auctionListDisplayer) IsInterfaceNil() bool {
	return ald == nil
}
