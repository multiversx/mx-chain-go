package dataRetriever

// HashSliceRequesterStub -
type HashSliceRequesterStub struct {
	RequesterStub
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
}

// RequestDataFromHashArray -
func (stub *HashSliceRequesterStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	if stub.RequestDataFromHashArrayCalled != nil {
		return stub.RequestDataFromHashArrayCalled(hashes, epoch)
	}
	return nil
}

// IsInterfaceNil -
func (stub *HashSliceRequesterStub) IsInterfaceNil() bool {
	return stub == nil
}
