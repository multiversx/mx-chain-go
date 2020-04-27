package disabled

// MessageSignVerifier represents the message verifier that accepts any message, sign, pk tuple
type MessageSignVerifier struct {
}

// Verify returns always nil
func (msv *MessageSignVerifier) Verify(_ []byte, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (msv *MessageSignVerifier) IsInterfaceNil() bool {
	return msv == nil
}
