package testscommon

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync/disabled"
)

type BootstrapperMock struct {
	process.Bootstrapper
}

// NewTestBootstrapperMock -
func NewTestBootstrapperMock() *BootstrapperMock {
	return &BootstrapperMock{
		Bootstrapper: disabled.NewDisabledBootstrapper(),
	}
}

// RollBack -
func (tbm *BootstrapperMock) RollBack(_ bool) error {
	return nil
}

// SetProbableHighestNonce -
func (tbm *BootstrapperMock) SetProbableHighestNonce(_ uint64) {
}
