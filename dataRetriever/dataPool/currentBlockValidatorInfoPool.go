package dataPool

import (
	"github.com/ElrondNetwork/elrond-go/state"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

var _ dataRetriever.ValidatorInfoCacher = (*validatorInfoMapCacher)(nil)

type validatorInfoMapCacher struct {
	mutValidatorInfo      sync.RWMutex
	validatorInfoForBlock map[string]*state.ShardValidatorInfo
}

// NewCurrentBlockValidatorInfoPool returns a new validator info pool to be used for the current block
func NewCurrentBlockValidatorInfoPool() *validatorInfoMapCacher {
	return &validatorInfoMapCacher{
		mutValidatorInfo:      sync.RWMutex{},
		validatorInfoForBlock: make(map[string]*state.ShardValidatorInfo),
	}
}

// Clean creates a new validator info pool
func (vimc *validatorInfoMapCacher) Clean() {
	vimc.mutValidatorInfo.Lock()
	vimc.validatorInfoForBlock = make(map[string]*state.ShardValidatorInfo)
	vimc.mutValidatorInfo.Unlock()
}

// GetValidatorInfo gets the validator info for the given hash
func (vimc *validatorInfoMapCacher) GetValidatorInfo(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
	vimc.mutValidatorInfo.RLock()
	defer vimc.mutValidatorInfo.RUnlock()

	validatorInfo, ok := vimc.validatorInfoForBlock[string(validatorInfoHash)]
	if !ok {
		return nil, dataRetriever.ErrValidatorInfoNotFoundInBlockPool
	}

	return validatorInfo, nil
}

// AddValidatorInfo adds the validator info for the given hash
func (vimc *validatorInfoMapCacher) AddValidatorInfo(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
	if check.IfNil(validatorInfo) {
		return
	}

	vimc.mutValidatorInfo.Lock()
	vimc.validatorInfoForBlock[string(validatorInfoHash)] = validatorInfo
	vimc.mutValidatorInfo.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (vimc *validatorInfoMapCacher) IsInterfaceNil() bool {
	return vimc == nil
}
