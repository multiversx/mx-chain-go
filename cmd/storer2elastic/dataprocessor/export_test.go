package dataprocessor

import "github.com/ElrondNetwork/elrond-go/data/state"

// SetPeerAdapter -
func (rp *ratingsProcessor) SetPeerAdapter(peerAdapter state.AccountsAdapter) {
	rp.peerAdapter = peerAdapter
}
