package genesis

// NilMessageSignVerifier represents the message verifier that accepts any message, sign, pk tuple
type NilMessageSignVerifier struct {
}

// Verify returns always nil
func (nmsv *NilMessageSignVerifier) Verify(_ []byte, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nmsv *NilMessageSignVerifier) IsInterfaceNil() bool {
	return nmsv == nil
}
