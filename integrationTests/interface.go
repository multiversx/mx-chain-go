package integrationTests

import (
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TestBootstrapper extends the Bootstrapper interface with some functions intended to be used only in tests
// as it simplifies the reproduction of edge cases
type TestBootstrapper interface {
	process.Bootstrapper
	RollBack(revertUsingForkNonce bool) error
	SetProbableHighestNonce(nonce uint64)
}

// TestEndOfEpochTrigger extends the endOfEpoch trigger interface with some functions intended to by used only
// in tests as it simplifies the reproduction of test scenarios
type TestEndOfEpochTrigger interface {
	endOfEpoch.TriggerHandler
	GetRoundsPerEpoch() int64
	SetTrigger(triggerHandler endOfEpoch.TriggerHandler)
	SetRoundsPerEpoch(roundsPerEpoch int64)
}
