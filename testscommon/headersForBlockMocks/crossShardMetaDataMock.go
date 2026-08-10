package headersForBlockMocks

// CrossShardMetaDataMock -
type CrossShardMetaDataMock struct {
	GetNonceCalled      func() uint64
	GetShardIdCalled    func() uint32
	GetHeaderHashCalled func() []byte
}

// GetNonce -
func (crossShardMetaDataMock *CrossShardMetaDataMock) GetNonce() uint64 {
	if crossShardMetaDataMock.GetNonceCalled != nil {
		return crossShardMetaDataMock.GetNonceCalled()
	}

	return 0
}

// GetShardID -
func (crossShardMetaDataMock *CrossShardMetaDataMock) GetShardID() uint32 {
	if crossShardMetaDataMock.GetShardIdCalled != nil {
		return crossShardMetaDataMock.GetShardIdCalled()
	}

	return 0
}

// GetHeaderHash -
func (crossShardMetaDataMock *CrossShardMetaDataMock) GetHeaderHash() []byte {
	if crossShardMetaDataMock.GetHeaderHashCalled != nil {
		return crossShardMetaDataMock.GetHeaderHashCalled()
	}
	return nil
}
