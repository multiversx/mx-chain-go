package dataRetriever

type TrieNodeRequesterStub struct {
	RequesterStub
	RequestDataFromHashArrayCalled                    func([][]byte, uint32) error
	RequestDataFromReferenceAndChunkCalled            func([]byte, uint32) error
	RequestDataFromHashArrayForRecoveryCalled         func([][]byte, uint32) error
	RequestDataFromReferenceAndChunkForRecoveryCalled func([]byte, uint32) error
}

func (stub *TrieNodeRequesterStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	if stub.RequestDataFromHashArrayCalled != nil {
		return stub.RequestDataFromHashArrayCalled(hashes, epoch)
	}
	return nil
}

func (stub *TrieNodeRequesterStub) RequestDataFromReferenceAndChunk(hash []byte, chunk uint32) error {
	if stub.RequestDataFromReferenceAndChunkCalled != nil {
		return stub.RequestDataFromReferenceAndChunkCalled(hash, chunk)
	}
	return nil
}

func (stub *TrieNodeRequesterStub) RequestDataFromHashArrayForRecovery(hashes [][]byte, epoch uint32) error {
	if stub.RequestDataFromHashArrayForRecoveryCalled != nil {
		return stub.RequestDataFromHashArrayForRecoveryCalled(hashes, epoch)
	}
	return nil
}

func (stub *TrieNodeRequesterStub) RequestDataFromReferenceAndChunkForRecovery(hash []byte, chunk uint32) error {
	if stub.RequestDataFromReferenceAndChunkForRecoveryCalled != nil {
		return stub.RequestDataFromReferenceAndChunkForRecoveryCalled(hash, chunk)
	}
	return nil
}

func (stub *TrieNodeRequesterStub) IsInterfaceNil() bool {
	return stub == nil
}
