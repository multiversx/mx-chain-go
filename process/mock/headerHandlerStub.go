package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type HeaderHandlerStub struct {
	GetMiniBlockHeadersWithDstCalled func(destId uint32) map[string]uint32
	GetPubKeysBitmapCalled           func() []byte
	GetSignatureCalled               func() []byte
	GetRootHashCalled                func() []byte
	GetRandSeedCalled                func() []byte
	GetPrevRandSeedCalled            func() []byte
	GetPrevHashCalled                func() []byte
	CloneCalled                      func() data.HeaderHandler
	GetChainIDCalled                 func() []byte
	CheckChainIDCalled               func(reference []byte) error
}

func (hhs *HeaderHandlerStub) Clone() data.HeaderHandler {
	return hhs.CloneCalled()
}

func (hhs *HeaderHandlerStub) IsStartOfEpochBlock() bool {
	return false
}

func (hhs *HeaderHandlerStub) GetShardID() uint32 {
	return 1
}

func (hhs *HeaderHandlerStub) SetShardID(shId uint32) {
}

func (hhs *HeaderHandlerStub) GetNonce() uint64 {
	return 1
}

func (hhs *HeaderHandlerStub) GetEpoch() uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetRound() uint64 {
	return 1
}

func (hhs *HeaderHandlerStub) GetTimeStamp() uint64 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetRootHash() []byte {
	return hhs.GetRootHashCalled()
}

func (hhs *HeaderHandlerStub) GetPrevHash() []byte {
	return hhs.GetPrevHashCalled()
}

func (hhs *HeaderHandlerStub) GetPrevRandSeed() []byte {
	return hhs.GetPrevRandSeedCalled()
}

func (hhs *HeaderHandlerStub) GetRandSeed() []byte {
	return hhs.GetRandSeedCalled()
}

func (hhs *HeaderHandlerStub) GetPubKeysBitmap() []byte {
	return hhs.GetPubKeysBitmapCalled()
}

func (hhs *HeaderHandlerStub) GetSignature() []byte {
	return hhs.GetSignatureCalled()
}

func (hhs *HeaderHandlerStub) GetLeaderSignature() []byte {
	return hhs.GetSignatureCalled()
}

func (hhs *HeaderHandlerStub) GetChainID() []byte {
	return hhs.GetChainIDCalled()
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
	return hhs.GetMiniBlockHeadersWithDstCalled(destId)
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
	return hhs.CheckChainIDCalled(reference)
}
