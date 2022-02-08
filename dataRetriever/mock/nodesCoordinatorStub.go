package mock

// NodesCoordinatorStub -
type NodesCoordinatorStub struct {
	GetAllEligibleValidatorsPublicKeysCalled func(epoch uint32) (map[uint32][][]byte, error)
}

// GetAllEligibleValidatorsPublicKeys -
func (nc *NodesCoordinatorStub) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	if nc.GetAllEligibleValidatorsPublicKeysCalled != nil {
		return nc.GetAllEligibleValidatorsPublicKeysCalled(epoch)
	}

	return nil, nil
}

// IsInterfaceNil -
func (nc *NodesCoordinatorStub) IsInterfaceNil() bool {
	return nc == nil
}
