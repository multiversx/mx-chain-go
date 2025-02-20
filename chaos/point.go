package chaos

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type PointInput struct {
	Name           string
	ConsensusState spos.ConsensusStateHandler
	NodePublicKey  string
	Header         data.HeaderHandler
	Signature      []byte
}
