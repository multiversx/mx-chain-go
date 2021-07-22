package bootstrapMocks

import (
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// EpochStartBootstrapperStub -
type EpochStartBootstrapperStub struct {
	TrieHolder      state.TriesHolder
	StorageManagers map[string]temporary.StorageManager
	BootstrapCalled func() (bootstrap.Parameters, error)
}

// GetTriesComponents -
func (esbs *EpochStartBootstrapperStub) GetTriesComponents() (state.TriesHolder, map[string]temporary.StorageManager) {
	return esbs.TrieHolder, esbs.StorageManagers
}

// Bootstrap -
func (esbs *EpochStartBootstrapperStub) Bootstrap() (bootstrap.Parameters, error) {
	if esbs.BootstrapCalled != nil {
		return esbs.BootstrapCalled()
	}

	return bootstrap.Parameters{}, nil
}

// EpochStartBootstrapperStub -
func (esbs *EpochStartBootstrapperStub) IsInterfaceNil() bool {
	return esbs == nil
}

// Close -
func (esbs *EpochStartBootstrapperStub) Close() error {
	return nil
}
