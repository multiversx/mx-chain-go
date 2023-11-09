package dataPool

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
)

var _ dataRetriever.ValidatorInfoCacher = (*validatorInfoMapCacher)(nil)

type validatorInfoMapCacher struct {
	mutValidatorInfo      sync.RWMutex
	validatorInfoForEpoch map[string]*state.ShardValidatorInfo
}

// NewCurrentEpochValidatorInfoPool returns a new validator info pool to be used for the current epoch
func NewCurrentEpochValidatorInfoPool() *validatorInfoMapCacher {
	return &validatorInfoMapCacher{
		mutValidatorInfo:      sync.RWMutex{},
		validatorInfoForEpoch: make(map[string]*state.ShardValidatorInfo),
	}
}

// Clean creates a new validator info pool
func (vimc *validatorInfoMapCacher) Clean() {
	vimc.mutValidatorInfo.Lock()
	vimc.validatorInfoForEpoch = make(map[string]*state.ShardValidatorInfo)
	vimc.mutValidatorInfo.Unlock()
}

// GetValidatorInfo gets the validator info for the given hash
func (vimc *validatorInfoMapCacher) GetValidatorInfo(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
	vimc.mutValidatorInfo.RLock()
	defer vimc.mutValidatorInfo.RUnlock()

	validatorInfo, ok := vimc.validatorInfoForEpoch[string(validatorInfoHash)]
	if !ok {
		return nil, dataRetriever.ErrValidatorInfoNotFoundInEpochPool
	}

	return validatorInfo, nil
}

// AddValidatorInfo adds the validator info for the given hash
func (vimc *validatorInfoMapCacher) AddValidatorInfo(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
	if check.IfNil(validatorInfo) {
		return
	}

	vimc.mutValidatorInfo.Lock()
	vimc.validatorInfoForEpoch[string(validatorInfoHash)] = validatorInfo
	vimc.mutValidatorInfo.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (vimc *validatorInfoMapCacher) IsInterfaceNil() bool {
	return vimc == nil
}
