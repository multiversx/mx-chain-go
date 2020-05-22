package mock

// BaseAccountMock -
type BaseAccountMock struct {
	AddressBytesField []byte
	Nonce             uint64
}

// AddressBytes -
func (bam *BaseAccountMock) AddressBytes() []byte {
	return bam.AddressBytesField
}

// IncreaseNonce -
func (bam *BaseAccountMock) IncreaseNonce(nonce uint64) {
	bam.Nonce += nonce
}

// GetNonce -
func (bam *BaseAccountMock) GetNonce() uint64 {
	return bam.Nonce
}

// IsInterfaceNil -
func (bam *BaseAccountMock) IsInterfaceNil() bool {
	return bam == nil
}
