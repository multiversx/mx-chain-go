package disabled

type multiSigner struct {
}

// NewMultiSigner returns a new instance of disabled multiSigner
func NewMultiSigner() *multiSigner {
	return &multiSigner{}
}

// CreateSignatureShare returns nil byte slice and nil error
func (m *multiSigner) CreateSignatureShare(_ []byte, _ []byte) ([]byte, error) {
	return nil, nil
}

// VerifySignatureShare returns nil
func (m *multiSigner) VerifySignatureShare(_ []byte, _ []byte, _ []byte) error {
	return nil
}

// AggregateSigs returns nil byte slice and nil error
func (m *multiSigner) AggregateSigs(_ [][]byte, _ [][]byte) ([]byte, error) {
	return nil, nil
}

// VerifyAggregatedSig returns nil
func (m *multiSigner) VerifyAggregatedSig(_ [][]byte, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *multiSigner) IsInterfaceNil() bool {
	return m == nil
}
