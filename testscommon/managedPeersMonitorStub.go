package testscommon

// ManagedPeersMonitorStub -
type ManagedPeersMonitorStub struct {
	GetManagedKeysCountCalled    func() int
	GetEligibleManagedKeysCalled func(epoch uint32) ([][]byte, error)
	GetWaitingManagedKeysCalled  func(epoch uint32) ([][]byte, error)
}

// GetManagedKeysCount -
func (stub *ManagedPeersMonitorStub) GetManagedKeysCount() int {
	if stub.GetManagedKeysCountCalled != nil {
		return stub.GetManagedKeysCountCalled()
	}
	return 0
}

// GetEligibleManagedKeys -
func (stub *ManagedPeersMonitorStub) GetEligibleManagedKeys(epoch uint32) ([][]byte, error) {
	if stub.GetEligibleManagedKeysCalled != nil {
		return stub.GetEligibleManagedKeysCalled(epoch)
	}
	return make([][]byte, 0), nil
}

// GetWaitingManagedKeys -
func (stub *ManagedPeersMonitorStub) GetWaitingManagedKeys(epoch uint32) ([][]byte, error) {
	if stub.GetWaitingManagedKeysCalled != nil {
		return stub.GetWaitingManagedKeysCalled(epoch)
	}
	return make([][]byte, 0), nil
}

func (stub *ManagedPeersMonitorStub) IsInterfaceNil() bool {
	return stub == nil
}
