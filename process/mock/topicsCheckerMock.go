package mock

// TopicsCheckerMock -
type TopicsCheckerMock struct {
	CheckValidityCalled func(topics [][]byte) error
}

// CheckValidity -
func (tc *TopicsCheckerMock) CheckValidity(topics [][]byte) error {
	if tc.CheckValidityCalled != nil {
		return tc.CheckValidityCalled(topics)
	}

	return nil
}

// IsInterfaceNil -
func (tc *TopicsCheckerMock) IsInterfaceNil() bool {
	return tc == nil
}
