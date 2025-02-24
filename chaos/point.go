package chaos

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type PointInput struct {
	Name           string
	ConsensusState spos.ConsensusStateHandler
	NodePublicKey  string
	Header         data.HeaderHandler
	Signature      []byte
	Corruptible    []interface{}
}

type PointOutput struct {
	HasValue  bool
	Error     error
	Boolean   bool
	Duration  time.Duration
	NumberInt int
}
