package testscommon

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/headerVersionData"
)

// HeaderHandlerStub -
type HeaderHandlerStub struct {
	EpochField                             uint32
	RoundField                             uint64
	TimestampField                         uint64
	BlockBodyTypeInt32Field                int32
	GetMiniBlockHeadersWithDstCalled       func(destId uint32) map[string]uint32
	GetOrderedCrossMiniblocksWithDstCalled func(destId uint32) []*data.MiniBlockInfo
	GetPubKeysBitmapCalled                 func() []byte
	GetSignatureCalled                     func() []byte
	GetRootHashCalled                      func() []byte
	GetRandSeedCalled                      func() []byte
	GetPrevRandSeedCalled                  func() []byte
	GetPrevHashCalled                      func() []byte
	CloneCalled                            func() data.HeaderHandler
	GetChainIDCalled                       func() []byte
	CheckChainIDCalled                     func(reference []byte) error
	GetReservedCalled                      func() []byte
	IsStartOfEpochBlockCalled              func() bool
	HasScheduledMiniBlocksCalled           func() bool
	GetNonceCalled                         func() uint64
	CheckFieldsForNilCalled                func() error
	SetShardIDCalled                       func(shardID uint32) error
	SetPrevHashCalled                      func(hash []byte) error
	SetPrevRandSeedCalled                  func(seed []byte) error
	SetPubKeysBitmapCalled                 func(bitmap []byte) error
	SetChainIDCalled                       func(chainID []byte) error
	SetTimeStampCalled                     func(timestamp uint64) error
	SetRandSeedCalled                      func(seed []byte) error
	SetSignatureCalled                     func(signature []byte) error
	SetLeaderSignatureCalled               func(signature []byte) error
}

// GetAccumulatedFees -
func (hhs *HeaderHandlerStub) GetAccumulatedFees() *big.Int {
	return big.NewInt(0)
}

// GetDeveloperFees -
func (hhs *HeaderHandlerStub) GetDeveloperFees() *big.Int {
	return big.NewInt(0)
}

// SetAccumulatedFees -
func (hhs *HeaderHandlerStub) SetAccumulatedFees(_ *big.Int) error {
	return nil
}

// SetDeveloperFees -
func (hhs *HeaderHandlerStub) SetDeveloperFees(_ *big.Int) error {
	return nil
}

// GetReceiptsHash -
func (hhs *HeaderHandlerStub) GetReceiptsHash() []byte {
	return []byte("receipt")
}

// SetShardID -
func (hhs *HeaderHandlerStub) SetShardID(shardID uint32) error {
	if hhs.SetShardIDCalled != nil {
		return hhs.SetShardIDCalled(shardID)
	}
	return nil
}

// IsStartOfEpochBlock -
func (hhs *HeaderHandlerStub) IsStartOfEpochBlock() bool {
	if hhs.IsStartOfEpochBlockCalled != nil {
		return hhs.IsStartOfEpochBlockCalled()
	}

	return false
}

// ShallowClone -
func (hhs *HeaderHandlerStub) ShallowClone() data.HeaderHandler {
	return hhs.CloneCalled()
}

// GetShardID -
func (hhs *HeaderHandlerStub) GetShardID() uint32 {
	return 1
}

// GetNonce -
func (hhs *HeaderHandlerStub) GetNonce() uint64 {
	if hhs.GetNonceCalled != nil {
		return hhs.GetNonceCalled()
	}
	return 1
}

// GetEpoch -
func (hhs *HeaderHandlerStub) GetEpoch() uint32 {
	return hhs.EpochField
}

// GetRound -
func (hhs *HeaderHandlerStub) GetRound() uint64 {
	return hhs.RoundField
}

// GetTimeStamp -
func (hhs *HeaderHandlerStub) GetTimeStamp() uint64 {
	return hhs.TimestampField
}

// GetRootHash -
func (hhs *HeaderHandlerStub) GetRootHash() []byte {
	return hhs.GetRootHashCalled()
}

// GetPrevHash -
func (hhs *HeaderHandlerStub) GetPrevHash() []byte {
	return hhs.GetPrevHashCalled()
}

// GetPrevRandSeed -
func (hhs *HeaderHandlerStub) GetPrevRandSeed() []byte {
	if hhs.GetPrevRandSeedCalled != nil {
		return hhs.GetPrevRandSeedCalled()
	}
	return make([]byte, 0)
}

// GetRandSeed -
func (hhs *HeaderHandlerStub) GetRandSeed() []byte {
	return hhs.GetRandSeedCalled()
}

// GetPubKeysBitmap -
func (hhs *HeaderHandlerStub) GetPubKeysBitmap() []byte {
	if hhs.GetPubKeysBitmapCalled != nil {
		return hhs.GetPubKeysBitmapCalled()
	}
	return make([]byte, 0)
}

// GetSignature -
func (hhs *HeaderHandlerStub) GetSignature() []byte {
	return hhs.GetSignatureCalled()
}

// GetLeaderSignature -
func (hhs *HeaderHandlerStub) GetLeaderSignature() []byte {
	return hhs.GetSignatureCalled()
}

// GetChainID -
func (hhs *HeaderHandlerStub) GetChainID() []byte {
	return hhs.GetChainIDCalled()
}

// GetTxCount -
func (hhs *HeaderHandlerStub) GetTxCount() uint32 {
	return 0
}

// GetReserved -
func (hhs *HeaderHandlerStub) GetReserved() []byte {
	if hhs.GetReservedCalled != nil {
		return hhs.GetReservedCalled()
	}

	return nil
}

// SetNonce -
func (hhs *HeaderHandlerStub) SetNonce(_ uint64) error {
	panic("implement me")
}

// SetEpoch -
func (hhs *HeaderHandlerStub) SetEpoch(_ uint32) error {
	panic("implement me")
}

// SetRound -
func (hhs *HeaderHandlerStub) SetRound(_ uint64) error {
	panic("implement me")
}

// SetTimeStamp -
func (hhs *HeaderHandlerStub) SetTimeStamp(timestamp uint64) error {
	if hhs.SetTimeStampCalled != nil {
		return hhs.SetTimeStampCalled(timestamp)
	}
	return nil
}

// SetRootHash -
func (hhs *HeaderHandlerStub) SetRootHash(_ []byte) error {
	panic("implement me")
}

// SetPrevHash -
func (hhs *HeaderHandlerStub) SetPrevHash(hash []byte) error {
	if hhs.SetPrevHashCalled != nil {
		return hhs.SetPrevHashCalled(hash)
	}
	return nil
}

// SetPrevRandSeed -
func (hhs *HeaderHandlerStub) SetPrevRandSeed(seed []byte) error {
	if hhs.SetPrevRandSeedCalled != nil {
		return hhs.SetPrevRandSeedCalled(seed)
	}
	return nil
}

// SetRandSeed -
func (hhs *HeaderHandlerStub) SetRandSeed(seed []byte) error {
	if hhs.SetRandSeedCalled != nil {
		return hhs.SetRandSeedCalled(seed)
	}
	return nil
}

// SetPubKeysBitmap -
func (hhs *HeaderHandlerStub) SetPubKeysBitmap(bitmap []byte) error {
	if hhs.SetPubKeysBitmapCalled != nil {
		return hhs.SetPubKeysBitmapCalled(bitmap)
	}
	return nil
}

// SetSignature -
func (hhs *HeaderHandlerStub) SetSignature(signature []byte) error {
	if hhs.SetSignatureCalled != nil {
		return hhs.SetSignatureCalled(signature)
	}
	return nil
}

// SetLeaderSignature -
func (hhs *HeaderHandlerStub) SetLeaderSignature(signature []byte) error {
	if hhs.SetLeaderSignatureCalled != nil {
		return hhs.SetLeaderSignatureCalled(signature)
	}
	return nil
}

// SetChainID -
func (hhs *HeaderHandlerStub) SetChainID(chainID []byte) error {
	if hhs.SetChainIDCalled != nil {
		return hhs.SetChainIDCalled(chainID)
	}
	return nil
}

// SetTxCount -
func (hhs *HeaderHandlerStub) SetTxCount(_ uint32) error {
	panic("implement me")
}

// GetMiniBlockHeadersWithDst -
func (hhs *HeaderHandlerStub) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	return hhs.GetMiniBlockHeadersWithDstCalled(destId)
}

// GetOrderedCrossMiniblocksWithDst -
func (hhs *HeaderHandlerStub) GetOrderedCrossMiniblocksWithDst(destId uint32) []*data.MiniBlockInfo {
	return hhs.GetOrderedCrossMiniblocksWithDstCalled(destId)
}

// GetMiniBlockHeadersHashes -
func (hhs *HeaderHandlerStub) GetMiniBlockHeadersHashes() [][]byte {
	panic("implement me")
}

// GetMiniBlockHeaderHandlers -
func (hhs *HeaderHandlerStub) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	panic("implement me")
}

// GetMetaBlockHashes -
func (hhs *HeaderHandlerStub) GetMetaBlockHashes() [][]byte {
	panic("implement me")
}

// GetBlockBodyTypeInt32 -
func (hhs *HeaderHandlerStub) GetBlockBodyTypeInt32() int32 {
	return hhs.BlockBodyTypeInt32Field
}

// GetValidatorStatsRootHash -
func (hhs *HeaderHandlerStub) GetValidatorStatsRootHash() []byte {
	return []byte("vs root hash")
}

// SetValidatorStatsRootHash -
func (hhs *HeaderHandlerStub) SetValidatorStatsRootHash(_ []byte) error {
	panic("implement me")
}

// SetMiniBlockHeaderHandlers -
func (hhs *HeaderHandlerStub) SetMiniBlockHeaderHandlers(_ []data.MiniBlockHeaderHandler) error {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (hhs *HeaderHandlerStub) IsInterfaceNil() bool {
	return hhs == nil
}

// GetEpochStartMetaHash -
func (hhs *HeaderHandlerStub) GetEpochStartMetaHash() []byte {
	panic("implement me")
}

// GetSoftwareVersion -
func (hhs *HeaderHandlerStub) GetSoftwareVersion() []byte {
	return []byte("softwareVersion")
}

// SetSoftwareVersion -
func (hhs *HeaderHandlerStub) SetSoftwareVersion(_ []byte) error {
	return nil
}

// SetReceiptsHash -
func (hhs *HeaderHandlerStub) SetReceiptsHash(_ []byte) error {
	return nil
}

// SetMetaBlockHashes -
func (hhs *HeaderHandlerStub) SetMetaBlockHashes(_ [][]byte) error {
	return nil
}

// SetEpochStartMetaHash -
func (hhs *HeaderHandlerStub) SetEpochStartMetaHash(_ []byte) error {
	return nil
}

// GetShardInfoHandlers -
func (hhs *HeaderHandlerStub) GetShardInfoHandlers() []data.ShardDataHandler {
	panic("implement me")
}

// GetEpochStartHandler -
func (hhs *HeaderHandlerStub) GetEpochStartHandler() data.EpochStartHandler {
	panic("implement me")
}

// GetDevFeesInEpoch -
func (hhs *HeaderHandlerStub) GetDevFeesInEpoch() *big.Int {
	panic("implement me")
}

// SetDevFeesInEpoch -
func (hhs *HeaderHandlerStub) SetDevFeesInEpoch(_ *big.Int) error {
	panic("implement me")
}

// SetShardInfoHandlers -
func (hhs *HeaderHandlerStub) SetShardInfoHandlers(_ []data.ShardDataHandler) error {
	panic("implement me")
}

// SetAccumulatedFeesInEpoch -
func (hhs *HeaderHandlerStub) SetAccumulatedFeesInEpoch(_ *big.Int) error {
	panic("implement me")
}

// SetScheduledRootHash -
func (hhs *HeaderHandlerStub) SetScheduledRootHash(_ []byte) error {
	panic("implement me")
}

// ValidateHeaderVersion -
func (hhs *HeaderHandlerStub) ValidateHeaderVersion() error {
	return nil
}

// SetAdditionalData sets the additional version-related data
func (hhs *HeaderHandlerStub) SetAdditionalData(_ headerVersionData.HeaderAdditionalData) error {
	return nil
}

// GetAdditionalData gets the additional version-related data
func (hhs *HeaderHandlerStub) GetAdditionalData() headerVersionData.HeaderAdditionalData {
	return nil
}

// HasScheduledSupport -
func (hhs *HeaderHandlerStub) HasScheduledSupport() bool {
	return false
}

// MapMiniBlockHashesToShards -
func (hhs *HeaderHandlerStub) MapMiniBlockHashesToShards() map[string]uint32 {
	panic("implement me")
}

// CheckFieldsForNil -
func (hhs *HeaderHandlerStub) CheckFieldsForNil() error {
	if hhs.CheckFieldsForNilCalled != nil {
		return hhs.CheckFieldsForNilCalled()
	}

	return nil
}

// HasScheduledMiniBlocks -
func (hhs *HeaderHandlerStub) HasScheduledMiniBlocks() bool {
	if hhs.HasScheduledMiniBlocksCalled != nil {
		return hhs.HasScheduledMiniBlocksCalled()
	}
	return false
}

// SetBlockBodyTypeInt32 -
func (hhs *HeaderHandlerStub) SetBlockBodyTypeInt32(blockBodyType int32) error {
	hhs.BlockBodyTypeInt32Field = blockBodyType

	return nil
}
