package factory

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

//HeaderSigVerifierHandler is the interface needed to check a header if is correct
type HeaderSigVerifierHandler interface {
	VerifyRandSeed(header data.HeaderHandler) error
	VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error
	VerifySignature(header data.HeaderHandler) error
	IsInterfaceNil() bool
}
