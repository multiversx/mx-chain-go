package testscommon

// ManagedPeersMonitorStub -
type ManagedPeersMonitorStub struct {
	GetManagedKeysCountCalled    func() int
	GetEligibleManagedKeysCalled func() ([][]byte, error)
	GetWaitingManagedKeysCalled  func() ([][]byte, error)
	GetManagedKeysCalled         func() [][]byte
	GetLoadedKeysCalled          func() [][]byte
}

// GetManagedKeys -
func (stub *ManagedPeersMonitorStub) GetManagedKeys() [][]byte {
	if stub.GetManagedKeysCalled != nil {
		return stub.GetManagedKeysCalled()
	}
	return make([][]byte, 0)
}

// GetLoadedKeys -
func (stub *ManagedPeersMonitorStub) GetLoadedKeys() [][]byte {
	if stub.GetLoadedKeysCalled != nil {
		return stub.GetLoadedKeysCalled()
	}
	return make([][]byte, 0)
}

// GetManagedKeysCount -
func (stub *ManagedPeersMonitorStub) GetManagedKeysCount() int {
	if stub.GetManagedKeysCountCalled != nil {
		return stub.GetManagedKeysCountCalled()
	}
	return 0
}

// GetEligibleManagedKeys -
func (stub *ManagedPeersMonitorStub) GetEligibleManagedKeys() ([][]byte, error) {
	if stub.GetEligibleManagedKeysCalled != nil {
		return stub.GetEligibleManagedKeysCalled()
	}
	return make([][]byte, 0), nil
}

// GetWaitingManagedKeys -
func (stub *ManagedPeersMonitorStub) GetWaitingManagedKeys() ([][]byte, error) {
	if stub.GetWaitingManagedKeysCalled != nil {
		return stub.GetWaitingManagedKeysCalled()
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil -
func (stub *ManagedPeersMonitorStub) IsInterfaceNil() bool {
	return stub == nil
}
