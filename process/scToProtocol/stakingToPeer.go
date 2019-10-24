package scToProtocol

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// stakingToPeer defines the component which will translate changes from staking SC state
// to validator statistics trie
type stakingToPeer struct {
	peerState state.AccountsAdapter
	baseState state.AccountsAdapter
}

func NewStakingToPeer() (*stakingToPeer, error) {
	return nil, nil
}

func UpdateProtocol(body data.BodyHandler) error {

}
