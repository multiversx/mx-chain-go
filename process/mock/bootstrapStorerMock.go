package mock

import "github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"

// BoostrapStorerMock -
type BoostrapStorerMock struct {
	PutCalled             func(round int64, bootData bootstrapStorage.BootstrapData) error
	GetCalled             func(round int64) (bootstrapStorage.BootstrapData, error)
	GetHighestRoundCalled func() int64
}

// Put -
func (bsm *BoostrapStorerMock) Put(round int64, bootData bootstrapStorage.BootstrapData) error {
	return bsm.PutCalled(round, bootData)
}

// Get -
func (bsm *BoostrapStorerMock) Get(round int64) (bootstrapStorage.BootstrapData, error) {
	return bsm.GetCalled(round)
}

// GetHighestRound -
func (bsm *BoostrapStorerMock) GetHighestRound() int64 {
	return bsm.GetHighestRoundCalled()
}

// SaveLastRound -
func (bsm *BoostrapStorerMock) SaveLastRound(round int64) error {
	return nil
}

// IsInterfaceNil -
func (bsm *BoostrapStorerMock) IsInterfaceNil() bool {
	return bsm == nil
}
