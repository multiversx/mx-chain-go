package headerForBlock

type nonceAndHashInfo struct {
	hash  []byte
	nonce uint64
}

// GetNonce returns the nonce
func (nahi *nonceAndHashInfo) GetNonce() uint64 {
	return nahi.nonce
}

// GetHash returns the hash
func (nahi *nonceAndHashInfo) GetHash() []byte {
	return nahi.hash
}

// IsInterfaceNil returns true if there is no value under the interface
func (nahi *nonceAndHashInfo) IsInterfaceNil() bool {
	return nahi == nil
}
