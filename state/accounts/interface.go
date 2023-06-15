package accounts

import (
	"github.com/multiversx/mx-chain-go/state"
)

type dataTrieInteractor interface {
	state.DataTrieTracker
}
