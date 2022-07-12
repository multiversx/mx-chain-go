package testscommon

// ProcessDebuggerStub -
type ProcessDebuggerStub struct {
	SetLastCommittedBlockRoundCalled func(round uint64)
	CloseCalled                      func() error
}

// SetLastCommittedBlockRound -
func (stub *ProcessDebuggerStub) SetLastCommittedBlockRound(round uint64) {
	if stub.SetLastCommittedBlockRoundCalled != nil {
		stub.SetLastCommittedBlockRoundCalled(round)
	}
}

// Close -
func (stub *ProcessDebuggerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *ProcessDebuggerStub) IsInterfaceNil() bool {
	return stub == nil
}
