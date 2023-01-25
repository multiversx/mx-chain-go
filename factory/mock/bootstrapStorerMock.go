package mock

import "github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"

// BootstrapStorerMock -
type BootstrapStorerMock struct {
	PutCalled             func(round int64, bootData bootstrapStorage.BootstrapData) error
	GetCalled             func(round int64) (bootstrapStorage.BootstrapData, error)
	GetHighestRoundCalled func() int64
}

// Put -
func (bsm *BootstrapStorerMock) Put(round int64, bootData bootstrapStorage.BootstrapData) error {
	return bsm.PutCalled(round, bootData)
}

// Get -
func (bsm *BootstrapStorerMock) Get(round int64) (bootstrapStorage.BootstrapData, error) {
	return bsm.GetCalled(round)
}

// GetHighestRound -
func (bsm *BootstrapStorerMock) GetHighestRound() int64 {
	if bsm.GetHighestRoundCalled != nil {
		return bsm.GetHighestRoundCalled()
	}

	return 0
}

// SaveLastRound -
func (bsm *BootstrapStorerMock) SaveLastRound(_ int64) error {
	return nil
}

// IsInterfaceNil -
func (bsm *BootstrapStorerMock) IsInterfaceNil() bool {
	return bsm == nil
}
