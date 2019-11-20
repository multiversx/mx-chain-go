package mock

type HeaderPoolsCleanerMock struct {
	CleanCalled             func(finalNonceInSelfShard uint64, finalNoncesInNotarizedShards map[uint32]uint64)
	NumRemovedHeadersCalled func() uint64
}

func (hpcm *HeaderPoolsCleanerMock) Clean(finalNonceInSelfShard uint64, finalNoncesInNotarizedShards map[uint32]uint64) {
	hpcm.CleanCalled(finalNonceInSelfShard, finalNoncesInNotarizedShards)
}

func (hpcm *HeaderPoolsCleanerMock) NumRemovedHeaders() uint64 {
	return hpcm.NumRemovedHeadersCalled()
}

func (hpcm *HeaderPoolsCleanerMock) IsInterfaceNil() bool {
	return hpcm == nil
}
