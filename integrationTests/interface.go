package integrationTests

import (
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TestBootstrapper extends the Bootstrapper interface with some functions intended to be used only in tests
// as it simplifies the reproduction of edge cases
type TestBootstrapper interface {
	process.Bootstrapper
	RollBack(revertUsingForkNonce bool) error
	SetProbableHighestNonce(nonce uint64)
}

// TestEpochStartTrigger extends the epochStart trigger interface with some functions intended to by used only
// in tests as it simplifies the reproduction of test scenarios
type TestEpochStartTrigger interface {
	epochStart.TriggerHandler
	GetRoundsPerEpoch() uint64
	SetTrigger(triggerHandler epochStart.TriggerHandler)
	SetRoundsPerEpoch(roundsPerEpoch uint64)
}
