package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type HeaderHandlerStub struct {
	GetPrevRandSeedCalled     func() []byte
	GetRandSeedCalled         func() []byte
	IsStartOfEpochBlockCalled func() bool
	GetEpochCaled             func() uint32
}

func (hhs *HeaderHandlerStub) GetReceiptsHash() []byte {
	return []byte("hash")
}

func (hhs *HeaderHandlerStub) Clone() data.HeaderHandler {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) IsStartOfEpochBlock() bool {
	if hhs.IsStartOfEpochBlockCalled != nil {
		return hhs.IsStartOfEpochBlockCalled()
	}

	return false
}

func (hhs *HeaderHandlerStub) GetShardID() uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetShardID(shId uint32) {
}

func (hhs *HeaderHandlerStub) GetNonce() uint64 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetEpoch() uint32 {
	if hhs.GetEpochCaled != nil {
		return hhs.GetEpochCaled()
	}

	return 0
}

func (hhs *HeaderHandlerStub) GetRound() uint64 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetTimeStamp() uint64 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetRootHash() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetPrevHash() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetPrevRandSeed() []byte {
	if hhs.GetPrevRandSeedCalled != nil {
		return hhs.GetPrevRandSeedCalled()
	}

	return []byte("prev rand seed")
}

func (hhs *HeaderHandlerStub) GetRandSeed() []byte {
	if hhs.GetRandSeedCalled != nil {
		return hhs.GetRandSeedCalled()
	}

	return []byte("rand seed")
}

func (hhs *HeaderHandlerStub) GetPubKeysBitmap() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetSignature() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetLeaderSignature() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetChainID() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetTxCount() uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetNonce(_ uint64) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetEpoch(_ uint32) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetRound(_ uint64) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetTimeStamp(_ uint64) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetRootHash(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetPrevHash(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetPrevRandSeed(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetRandSeed(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetPubKeysBitmap(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetSignature(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetLeaderSignature(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetChainID(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetTxCount(_ uint32) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetValidatorStatsRootHash() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetValidatorStatsRootHash(_ []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetMiniBlockProcessed(_ []byte) bool {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetMiniBlockProcessed(_ []byte, _ bool) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (hhs *HeaderHandlerStub) IsInterfaceNil() bool {
	return hhs == nil
}

func (hhs *HeaderHandlerStub) ItemsInHeader() uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) ItemsInBody() uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) CheckChainID(reference []byte) error {
	panic("implement me")
}
