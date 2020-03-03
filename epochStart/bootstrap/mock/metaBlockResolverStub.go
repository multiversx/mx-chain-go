package mock

// MetaBlockResolverStub -
type MetaBlockResolverStub struct {
	RequestEpochStartMetaBlockCalled func(epoch uint32) error
}

// RequestEpochStartMetaBlock -
func (m *MetaBlockResolverStub) RequestEpochStartMetaBlock(epoch uint32) error {
	if m.RequestEpochStartMetaBlockCalled != nil {
		return m.RequestEpochStartMetaBlockCalled(epoch)
	}

	return nil
}

// IsInterfaceNil -
func (m *MetaBlockResolverStub) IsInterfaceNil() bool {
	return m == nil
}
