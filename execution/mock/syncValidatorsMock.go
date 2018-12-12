package mock

import (
	"math/big"
)

type SyncValidatorsMock struct {
	AddValidatorCalled    func(nodeId string, stake big.Int) error
	RemoveValidatorCalled func(nodeId string) error
}

func (svm *SyncValidatorsMock) AddValidator(nodeId string, stake big.Int) {
	svm.AddValidatorCalled(nodeId, stake)
}

func (svm *SyncValidatorsMock) RemoveValidator(nodeId string) {
	svm.RemoveValidatorCalled(nodeId)
}
