package mock

// NodesSetupMock -
type NodesSetupMock struct {
	MinShardHysteresisNodesCalled func() uint32
	MinMetaHysteresisNodesCalled  func() uint32
}

// MinShardHysteresisNodes -
func (mock *NodesSetupMock) MinShardHysteresisNodes() uint32 {
	if mock.MinShardHysteresisNodesCalled != nil {
		return mock.MinShardHysteresisNodesCalled()
	}
	return 1
}

// MinMetaHysteresisNodes -
func (mock *NodesSetupMock) MinMetaHysteresisNodes() uint32 {
	if mock.MinMetaHysteresisNodesCalled != nil {
		return mock.MinMetaHysteresisNodesCalled()
	}
	return 1
}

// IsInterfaceNil -
func (mock *NodesSetupMock) IsInterfaceNil() bool {
	return mock == nil
}
