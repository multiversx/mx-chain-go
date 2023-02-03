package mock

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// InternalBlockApiHandlerStub -
type InternalBlockApiHandlerStub struct {
	GetInternalShardBlockByNonceCalled          func(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHashCalled           func(format common.ApiOutputFormat, hash []byte) (interface{}, error)
	GetInternalShardBlockByRoundCalled          func(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonceCalled           func(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHashCalled            func(format common.ApiOutputFormat, hash []byte) (interface{}, error)
	GetInternalMetaBlockByRoundCalled           func(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMiniBlockCalled                  func(format common.ApiOutputFormat, hash []byte, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochMetaBlockCalled      func(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochValidatorsInfoCalled func(epoch uint32) ([]*state.ShardValidatorInfo, error)
}

// GetInternalShardBlockByNonce -
func (ibah *InternalBlockApiHandlerStub) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if ibah.GetInternalShardBlockByNonceCalled != nil {
		return ibah.GetInternalShardBlockByNonceCalled(format, nonce)
	}
	return nil, nil
}

// GetInternalShardBlockByHash -
func (ibah *InternalBlockApiHandlerStub) GetInternalShardBlockByHash(format common.ApiOutputFormat, hash []byte) (interface{}, error) {
	if ibah.GetInternalShardBlockByHashCalled != nil {
		return ibah.GetInternalShardBlockByHashCalled(format, hash)
	}
	return nil, nil
}

// GetInternalShardBlockByRound -
func (ibah *InternalBlockApiHandlerStub) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if ibah.GetInternalShardBlockByRoundCalled != nil {
		return ibah.GetInternalShardBlockByRoundCalled(format, round)
	}
	return nil, nil
}

// GetInternalMetaBlockByNonce -
func (ibah *InternalBlockApiHandlerStub) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if ibah.GetInternalMetaBlockByNonceCalled != nil {
		return ibah.GetInternalMetaBlockByNonceCalled(format, nonce)
	}
	return nil, nil
}

// GetInternalMetaBlockByHash -
func (ibah *InternalBlockApiHandlerStub) GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash []byte) (interface{}, error) {
	if ibah.GetInternalMetaBlockByHashCalled != nil {
		return ibah.GetInternalMetaBlockByHashCalled(format, hash)
	}
	return nil, nil
}

// GetInternalMetaBlockByRound -
func (ibah *InternalBlockApiHandlerStub) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if ibah.GetInternalMetaBlockByRoundCalled != nil {
		return ibah.GetInternalMetaBlockByRoundCalled(format, round)
	}
	return nil, nil
}

// GetInternalMiniBlock -
func (ibah *InternalBlockApiHandlerStub) GetInternalMiniBlock(format common.ApiOutputFormat, hash []byte, epoch uint32) (interface{}, error) {
	if ibah.GetInternalMiniBlockCalled != nil {
		return ibah.GetInternalMiniBlockCalled(format, hash, epoch)
	}
	return nil, nil
}

// GetInternalStartOfEpochMetaBlock -
func (ibah *InternalBlockApiHandlerStub) GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error) {
	if ibah.GetInternalStartOfEpochMetaBlockCalled != nil {
		return ibah.GetInternalStartOfEpochMetaBlockCalled(format, epoch)
	}
	return nil, nil
}

// GetInternalStartOfEpochValidatorsInfo -
func (ibah *InternalBlockApiHandlerStub) GetInternalStartOfEpochValidatorsInfo(epoch uint32) ([]*state.ShardValidatorInfo, error) {
	if ibah.GetInternalStartOfEpochValidatorsInfoCalled != nil {
		return ibah.GetInternalStartOfEpochValidatorsInfoCalled(epoch)
	}
	return nil, nil
}

// IsInterfaceNil -
func (ibah *InternalBlockApiHandlerStub) IsInterfaceNil() bool {
	return ibah == nil
}
