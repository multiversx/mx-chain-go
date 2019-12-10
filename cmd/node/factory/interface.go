package factory

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

//HeaderSigVerifierHandler is the interface needed to check a header if is correct
type HeaderSigVerifierHandler interface {
	VerifyRandSeed(header data.HeaderHandler) error
	process.InterceptedHeaderSigVerifier
}
