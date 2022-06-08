package testscommon

type AccountBlockInfoStub struct {
	Nonce    uint64
	Hash     []byte
	RootHash []byte
}

// GetNonce returns the block nonce
func (stub *AccountBlockInfoStub) GetNonce() uint64 {
	return stub.Nonce
}

// GetHash returns the block hash
func (stub *AccountBlockInfoStub) GetHash() []byte {
	return stub.Hash
}

// GetRootHash returns the block rootHash
func (stub *AccountBlockInfoStub) GetRootHash() []byte {
	return stub.RootHash
}

// IsInterfaceNil -
func (stub *AccountBlockInfoStub) IsInterfaceNil() bool {
	return stub == nil
}
