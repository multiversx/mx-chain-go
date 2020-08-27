package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// HeaderHandlerStub --
type HeaderHandlerStub struct {
	GetPrevRandSeedCalled     func() []byte
	GetRandSeedCalled         func() []byte
	IsStartOfEpochBlockCalled func() bool
	GetEpochCaled             func() uint32
}

// GetAccumulatedFees --
func (hhs *HeaderHandlerStub) GetAccumulatedFees() *big.Int {
	panic("implement me")
}

// GetDeveloperFees --
func (hhs *HeaderHandlerStub) GetDeveloperFees() *big.Int {
	panic("implement me")
}

// SetAccumulatedFees --
func (hhs *HeaderHandlerStub) SetAccumulatedFees(_ *big.Int) {
	panic("implement me")
}

// SetDeveloperFees --
func (hhs *HeaderHandlerStub) SetDeveloperFees(_ *big.Int) {
	panic("implement me")
}

// GetReceiptsHash --
func (hhs *HeaderHandlerStub) GetReceiptsHash() []byte {
	return []byte("hash")
}

// Clone --
func (hhs *HeaderHandlerStub) Clone() data.HeaderHandler {
	panic("implement me")
}

// IsStartOfEpochBlock --
func (hhs *HeaderHandlerStub) IsStartOfEpochBlock() bool {
	if hhs.IsStartOfEpochBlockCalled != nil {
		return hhs.IsStartOfEpochBlockCalled()
	}

	return false
}

// GetShardID --
func (hhs *HeaderHandlerStub) GetShardID() uint32 {
	panic("implement me")
}

// SetShardID --
func (hhs *HeaderHandlerStub) SetShardID(_ uint32) {
}

// GetNonce -
func (hhs *HeaderHandlerStub) GetNonce() uint64 {
	panic("implement me")
}

// GetEpoch -
func (hhs *HeaderHandlerStub) GetEpoch() uint32 {
	if hhs.GetEpochCaled != nil {
		return hhs.GetEpochCaled()
	}

	return 0
}

// GetRound -
func (hhs *HeaderHandlerStub) GetRound() uint64 {
	panic("implement me")
}

// GetTimeStamp -
func (hhs *HeaderHandlerStub) GetTimeStamp() uint64 {
	panic("implement me")
}

// GetRootHash -
func (hhs *HeaderHandlerStub) GetRootHash() []byte {
	panic("implement me")
}

// GetPrevHash -
func (hhs *HeaderHandlerStub) GetPrevHash() []byte {
	panic("implement me")
}

// GetPrevRandSeed -
func (hhs *HeaderHandlerStub) GetPrevRandSeed() []byte {
	if hhs.GetPrevRandSeedCalled != nil {
		return hhs.GetPrevRandSeedCalled()
	}

	return []byte("prev rand seed")
}

// GetRandSeed -
func (hhs *HeaderHandlerStub) GetRandSeed() []byte {
	if hhs.GetRandSeedCalled != nil {
		return hhs.GetRandSeedCalled()
	}

	return []byte("rand seed")
}

// GetPubKeysBitmap -
func (hhs *HeaderHandlerStub) GetPubKeysBitmap() []byte {
	panic("implement me")
}

// GetSignature -
func (hhs *HeaderHandlerStub) GetSignature() []byte {
	panic("implement me")
}

// GetLeaderSignature -
func (hhs *HeaderHandlerStub) GetLeaderSignature() []byte {
	panic("implement me")
}

// GetChainID -
func (hhs *HeaderHandlerStub) GetChainID() []byte {
	panic("implement me")
}

// GetTxCount -
func (hhs *HeaderHandlerStub) GetTxCount() uint32 {
	panic("implement me")
}

// GetReserved -
func (hhs *HeaderHandlerStub) GetReserved() []byte {
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
func (hhs *HeaderHandlerStub) GetMiniBlockHeadersWithDst(_ uint32) map[string]uint32 {
	panic("implement me")
}

// GetMiniBlockHeadersHashes -
func (hhs *HeaderHandlerStub) GetMiniBlockHeadersHashes() [][]byte {
	panic("implement me")
}

// GetValidatorStatsRootHash -
func (hhs *HeaderHandlerStub) GetValidatorStatsRootHash() []byte {
	panic("implement me")
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
	return []byte("v1.0")
}

// SetSoftwareVersion -
func (hhs *HeaderHandlerStub) SetSoftwareVersion(_ []byte) {
}
