package epochStart

import "github.com/multiversx/mx-chain-core-go/data"

// EpochStartShardDataStub -
type EpochStartShardDataStub struct {
	GetShardIDCalled                        func() uint32
	GetEpochCalled                          func() uint32
	GetRoundCalled                          func() uint64
	GetNonceCalled                          func() uint64
	GetHeaderHashCalled                     func() []byte
	GetRootHashCalled                       func() []byte
	GetFirstPendingMetaBlockCalled          func() []byte
	GetLastFinishedMetaBlockCalled          func() []byte
	GetPendingMiniBlockHeaderHandlersCalled func() []data.MiniBlockHeaderHandler
	SetShardIDCalled                        func(uint32) error
	SetEpochCalled                          func(uint32) error
	SetRoundCalled                          func(uint64) error
	SetNonceCalled                          func(uint64) error
	SetHeaderHashCalled                     func([]byte) error
	SetRootHashCalled                       func([]byte) error
	SetFirstPendingMetaBlockCalled          func([]byte) error
	SetLastFinishedMetaBlockCalled          func([]byte) error
	SetPendingMiniBlockHeadersCalled        func([]data.MiniBlockHeaderHandler) error
}

// GetShardID -
func (essds *EpochStartShardDataStub) GetShardID() uint32 {
	if essds.GetShardIDCalled != nil {
		return essds.GetShardIDCalled()
	}

	return 0
}

// GetEpoch -
func (essds *EpochStartShardDataStub) GetEpoch() uint32 {
	if essds.GetEpochCalled != nil {
		return essds.GetEpochCalled()
	}
	return 0
}

// GetRound -
func (essds *EpochStartShardDataStub) GetRound() uint64 {
	if essds.GetRoundCalled != nil {
		return essds.GetRoundCalled()
	}
	return 0
}

// GetNonce -
func (essds *EpochStartShardDataStub) GetNonce() uint64 {
	if essds.GetNonceCalled != nil {
		return essds.GetNonceCalled()
	}
	return 0
}

// GetHeaderHash -
func (essds *EpochStartShardDataStub) GetHeaderHash() []byte {
	if essds.GetHeaderHashCalled != nil {
		return essds.GetHeaderHashCalled()
	}
	return []byte("header hash")
}

// GetRootHash -
func (essds *EpochStartShardDataStub) GetRootHash() []byte {
	if essds.GetRootHashCalled != nil {
		return essds.GetRootHashCalled()
	}
	return []byte("root hash")
}

// GetFirstPendingMetaBlock -
func (essds *EpochStartShardDataStub) GetFirstPendingMetaBlock() []byte {
	if essds.GetFirstPendingMetaBlockCalled != nil {
		return essds.GetFirstPendingMetaBlockCalled()
	}
	return nil
}

// GetLastFinishedMetaBlock -
func (essds *EpochStartShardDataStub) GetLastFinishedMetaBlock() []byte {
	if essds.GetLastFinishedMetaBlockCalled != nil {
		return essds.GetLastFinishedMetaBlockCalled()
	}
	return nil
}

// GetPendingMiniBlockHeaderHandlers -
func (essds *EpochStartShardDataStub) GetPendingMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if essds.GetPendingMiniBlockHeaderHandlersCalled != nil {
		return essds.GetPendingMiniBlockHeaderHandlersCalled()
	}
	return nil
}

// SetShardID -
func (essds *EpochStartShardDataStub) SetShardID(shardID uint32) error {
	if essds.SetShardIDCalled != nil {
		return essds.SetShardIDCalled(shardID)
	}
	return nil
}

// SetEpoch -
func (essds *EpochStartShardDataStub) SetEpoch(epoch uint32) error {
	if essds.SetEpochCalled != nil {
		return essds.SetEpochCalled(epoch)
	}
	return nil
}

// SetRound -
func (essds *EpochStartShardDataStub) SetRound(round uint64) error {
	if essds.SetRoundCalled != nil {
		return essds.SetRoundCalled(round)
	}
	return nil
}

// SetNonce -
func (essds *EpochStartShardDataStub) SetNonce(nonce uint64) error {
	if essds.SetNonceCalled != nil {
		return essds.SetNonceCalled(nonce)
	}
	return nil
}

// SetHeaderHash -
func (essds *EpochStartShardDataStub) SetHeaderHash(hash []byte) error {
	if essds.SetHeaderHashCalled != nil {
		return essds.SetHeaderHashCalled(hash)
	}
	return nil
}

// SetRootHash -
func (essds *EpochStartShardDataStub) SetRootHash(rootHash []byte) error {
	if essds.SetRootHashCalled != nil {
		return essds.SetRootHashCalled(rootHash)
	}
	return nil
}

// SetFirstPendingMetaBlock -
func (essds *EpochStartShardDataStub) SetFirstPendingMetaBlock(metaBlock []byte) error {
	if essds.SetFirstPendingMetaBlockCalled != nil {
		return essds.SetFirstPendingMetaBlockCalled(metaBlock)
	}
	return nil
}

// SetLastFinishedMetaBlock -
func (essds *EpochStartShardDataStub) SetLastFinishedMetaBlock(metaBlock []byte) error {
	if essds.SetLastFinishedMetaBlockCalled != nil {
		return essds.SetLastFinishedMetaBlockCalled(metaBlock)
	}
	return nil
}

// SetPendingMiniBlockHeaders -
func (essds *EpochStartShardDataStub) SetPendingMiniBlockHeaders(mbHeaders []data.MiniBlockHeaderHandler) error {
	if essds.SetPendingMiniBlockHeadersCalled != nil {
		return essds.SetPendingMiniBlockHeadersCalled(mbHeaders)
	}
	return nil
}
