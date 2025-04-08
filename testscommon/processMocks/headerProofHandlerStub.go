package processMocks

// HeaderProofHandlerStub -
type HeaderProofHandlerStub struct {
	GetPubKeysBitmapCalled       func() []byte
	GetAggregatedSignatureCalled func() []byte
	GetHeaderHashCalled          func() []byte
	GetHeaderEpochCalled         func() uint32
	GetHeaderNonceCalled         func() uint64
	GetHeaderShardIdCalled       func() uint32
	GetHeaderRoundCalled         func() uint64
	GetIsStartOfEpochCalled      func() bool
}

// GetPubKeysBitmap -
func (h *HeaderProofHandlerStub) GetPubKeysBitmap() []byte {
	if h.GetPubKeysBitmapCalled != nil {
		return h.GetPubKeysBitmapCalled()
	}
	return nil
}

// GetAggregatedSignature -
func (h *HeaderProofHandlerStub) GetAggregatedSignature() []byte {
	if h.GetAggregatedSignatureCalled != nil {
		return h.GetAggregatedSignatureCalled()
	}
	return nil
}

// GetHeaderHash -
func (h *HeaderProofHandlerStub) GetHeaderHash() []byte {
	if h.GetHeaderHashCalled != nil {
		return h.GetHeaderHashCalled()
	}
	return nil
}

// GetHeaderEpoch -
func (h *HeaderProofHandlerStub) GetHeaderEpoch() uint32 {
	if h.GetHeaderEpochCalled != nil {
		return h.GetHeaderEpochCalled()
	}
	return 0
}

// GetHeaderNonce -
func (h *HeaderProofHandlerStub) GetHeaderNonce() uint64 {
	if h.GetHeaderNonceCalled != nil {
		return h.GetHeaderNonceCalled()
	}
	return 0
}

// GetHeaderShardId -
func (h *HeaderProofHandlerStub) GetHeaderShardId() uint32 {
	if h.GetHeaderShardIdCalled != nil {
		return h.GetHeaderShardIdCalled()
	}
	return 0
}

// GetHeaderRound -
func (h *HeaderProofHandlerStub) GetHeaderRound() uint64 {
	if h.GetHeaderRoundCalled != nil {
		return h.GetHeaderRoundCalled()
	}
	return 0
}

// GetIsStartOfEpoch -
func (h *HeaderProofHandlerStub) GetIsStartOfEpoch() bool {
	if h.GetIsStartOfEpochCalled != nil {
		return h.GetIsStartOfEpochCalled()
	}
	return false
}

// IsInterfaceNil -
func (h *HeaderProofHandlerStub) IsInterfaceNil() bool {
	return h == nil
}
