package testscommon

// PeerAuthenticationPayloadValidatorStub -
type PeerAuthenticationPayloadValidatorStub struct {
	ValidateTimestampCalled func(payloadTimestamp int64) error
}

// ValidateTimestamp -
func (stub *PeerAuthenticationPayloadValidatorStub) ValidateTimestamp(payloadTimestamp int64) error {
	if stub.ValidateTimestampCalled != nil {
		return stub.ValidateTimestampCalled(payloadTimestamp)
	}

	return nil
}

// IsInterfaceNil -
func (stub *PeerAuthenticationPayloadValidatorStub) IsInterfaceNil() bool {
	return stub == nil
}
