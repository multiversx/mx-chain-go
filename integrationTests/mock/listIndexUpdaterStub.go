package mock

// ListIndexUpdaterStub -
type ListIndexUpdaterStub struct {
	UpdateListAndIndexCalled func(pubKey string, list string, index int) error
}

// UpdateListAndIndex -
func (lius *ListIndexUpdaterStub) UpdateListAndIndex(pubKey string, list string, index int) error {
	if lius.UpdateListAndIndexCalled != nil {
		return lius.UpdateListAndIndexCalled(pubKey, list, index)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (lius *ListIndexUpdaterStub) IsInterfaceNil() bool {
	return lius == nil
}
