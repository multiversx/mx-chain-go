package disabled

// DisabledMessageSignVerifier represents the message verifier that accepts any message, sign, pk tuple
type DisabledMessageSignVerifier struct {
}

// Verify returns always nil
func (dmsv *DisabledMessageSignVerifier) Verify(_ []byte, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dmsv *DisabledMessageSignVerifier) IsInterfaceNil() bool {
	return dmsv == nil
}
