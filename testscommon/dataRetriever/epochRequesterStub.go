package dataRetriever

// EpochRequesterStub -
type EpochRequesterStub struct {
	RequesterStub
	RequestDataFromEpochCalled func(identifier []byte) error
}

// RequestDataFromEpoch -
func (stub *EpochRequesterStub) RequestDataFromEpoch(identifier []byte) error {
	if stub.RequestDataFromEpochCalled != nil {
		return stub.RequestDataFromEpochCalled(identifier)
	}
	return nil
}
