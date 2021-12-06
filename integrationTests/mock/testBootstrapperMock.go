package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/sync/disabled"
)

type testBootstrapperMock struct {
	process.Bootstrapper
}

// NewTestBootstrapperMock -
func NewTestBootstrapperMock() *testBootstrapperMock {
	return &testBootstrapperMock{
		Bootstrapper: disabled.NewDisabledBootstrapper(),
	}
}

// RollBack -
func (tbm *testBootstrapperMock) RollBack(_ bool) error {
	return nil
}

// SetProbableHighestNonce -
func (tbm *testBootstrapperMock) SetProbableHighestNonce(_ uint64) {
}
