package mock

import "github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"

// BoostrapStorerStub -
type BoostrapStorerStub struct {
	PutCalled             func(round int64, bootData bootstrapStorage.BootstrapData) error
	GetCalled             func(round int64) (bootstrapStorage.BootstrapData, error)
	GetHighestRoundCalled func() int64
}

// Put -
func (bsm *BoostrapStorerStub) Put(round int64, bootData bootstrapStorage.BootstrapData) error {
	return bsm.PutCalled(round, bootData)
}

// Get -
func (bsm *BoostrapStorerStub) Get(round int64) (bootstrapStorage.BootstrapData, error) {
	return bsm.GetCalled(round)
}

// GetHighestRound -
func (bsm *BoostrapStorerStub) GetHighestRound() int64 {
	return bsm.GetHighestRoundCalled()
}

// SaveLastRound -
func (bsm *BoostrapStorerStub) SaveLastRound(_ int64) error {
	return nil
}

// IsInterfaceNil -
func (bsm *BoostrapStorerStub) IsInterfaceNil() bool {
	return bsm == nil
}
