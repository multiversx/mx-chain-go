package mock

type HeaderHandlerStub struct {
	GetMiniBlockHeadersWithDstCalled func(destId uint32) map[string]uint32
}

func (hhs *HeaderHandlerStub) GetShardID() uint32 {
	return 1
}

func (hhs *HeaderHandlerStub) GetNonce() uint64 {
	return 1
}

func (hhs *HeaderHandlerStub) GetEpoch() uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetRound() uint32 {
	return 1
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
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetRandSeed() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetPubKeysBitmap() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetSignature() []byte {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetTxCount() uint32 {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetNonce(n uint64) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetEpoch(e uint32) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetRound(r uint32) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetTimeStamp(ts uint64) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetRootHash(rHash []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetPrevHash(pvHash []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetPrevRandSeed(pvRandSeed []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetRandSeed(randSeed []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetPubKeysBitmap(pkbm []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetSignature(sg []byte) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetTxCount(txCount uint32) {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	return hhs.GetMiniBlockHeadersWithDstCalled(destId)
}

func (hhs *HeaderHandlerStub) GetMiniBlockProcessed(hash []byte) bool {
	panic("implement me")
}

func (hhs *HeaderHandlerStub) SetMiniBlockProcessed(hash []byte, processed bool) {
	panic("implement me")
}
