package mock

import "github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"

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
	if bsm.GetCalled == nil {
		return bootstrapStorage.BootstrapData{}, bootstrapStorage.ErrNilMarshalizer
	}
	return bsm.GetCalled(round)
}

// GetHighestRound -
func (bsm *BoostrapStorerMock) GetHighestRound() int64 {
	if bsm.GetHighestRoundCalled == nil {
		return 0
	}
	return bsm.GetHighestRoundCalled()
}

// IsInterfaceNil -
func (bsm *BoostrapStorerMock) IsInterfaceNil() bool {
	return bsm == nil
}

// SaveLastRound -
func (bsm *BoostrapStorerMock) SaveLastRound(_ int64) error {
	return nil
}
