package testscommon

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

// HeaderHandlerStub -
type HeaderHandlerStub struct {
	EpochField                             uint32
	TimestampField                         uint64
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
func (hhs *HeaderHandlerStub) SetAccumulatedFees(_ *big.Int) {
}

// SetDeveloperFees -
func (hhs *HeaderHandlerStub) SetDeveloperFees(_ *big.Int) {
}

// GetReceiptsHash -
func (hhs *HeaderHandlerStub) GetReceiptsHash() []byte {
	return []byte("receipt")
}

// SetShardID -
func (hhs *HeaderHandlerStub) SetShardID(_ uint32) {
}

// IsStartOfEpochBlock -
func (hhs *HeaderHandlerStub) IsStartOfEpochBlock() bool {
	if hhs.IsStartOfEpochBlockCalled != nil {
		return hhs.IsStartOfEpochBlockCalled()
	}

	return false
}

// Clone -
func (hhs *HeaderHandlerStub) Clone() data.HeaderHandler {
	return hhs.CloneCalled()
}

// GetShardID -
func (hhs *HeaderHandlerStub) GetShardID() uint32 {
	return 1
}

// GetNonce -
func (hhs *HeaderHandlerStub) GetNonce() uint64 {
	return 1
}

// GetEpoch -
func (hhs *HeaderHandlerStub) GetEpoch() uint32 {
	return hhs.EpochField
}

// GetRound -
func (hhs *HeaderHandlerStub) GetRound() uint64 {
	return 1
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
	return hhs.GetPrevRandSeedCalled()
}

// GetRandSeed -
func (hhs *HeaderHandlerStub) GetRandSeed() []byte {
	return hhs.GetRandSeedCalled()
}

// GetPubKeysBitmap -
func (hhs *HeaderHandlerStub) GetPubKeysBitmap() []byte {
	return hhs.GetPubKeysBitmapCalled()
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
func (hhs *HeaderHandlerStub) SetNonce(_ uint64) {
	panic("implement me")
}

// SetEpoch -
func (hhs *HeaderHandlerStub) SetEpoch(_ uint32) {
	panic("implement me")
}

// SetRound -
func (hhs *HeaderHandlerStub) SetRound(_ uint64) {
	panic("implement me")
}

// SetTimeStamp -
func (hhs *HeaderHandlerStub) SetTimeStamp(_ uint64) {
	panic("implement me")
}

// SetRootHash -
func (hhs *HeaderHandlerStub) SetRootHash(_ []byte) {
	panic("implement me")
}

// SetPrevHash -
func (hhs *HeaderHandlerStub) SetPrevHash(_ []byte) {
	panic("implement me")
}

// SetPrevRandSeed -
func (hhs *HeaderHandlerStub) SetPrevRandSeed(_ []byte) {
	panic("implement me")
}

// SetRandSeed -
func (hhs *HeaderHandlerStub) SetRandSeed(_ []byte) {
	panic("implement me")
}

// SetPubKeysBitmap -
func (hhs *HeaderHandlerStub) SetPubKeysBitmap(_ []byte) {
	panic("implement me")
}

// SetSignature -
func (hhs *HeaderHandlerStub) SetSignature(_ []byte) {
	panic("implement me")
}

// SetLeaderSignature -
func (hhs *HeaderHandlerStub) SetLeaderSignature(_ []byte) {
	panic("implement me")
}

// SetChainID -
func (hhs *HeaderHandlerStub) SetChainID(_ []byte) {
	panic("implement me")
}

// SetTxCount -
func (hhs *HeaderHandlerStub) SetTxCount(_ uint32) {
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

// GetValidatorStatsRootHash -
func (hhs *HeaderHandlerStub) GetValidatorStatsRootHash() []byte {
	return []byte("vs root hash")
}

// SetValidatorStatsRootHash -
func (hhs *HeaderHandlerStub) SetValidatorStatsRootHash(_ []byte) {
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
func (hhs *HeaderHandlerStub) SetSoftwareVersion(_ []byte) {
}
