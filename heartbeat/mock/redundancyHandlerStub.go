package mock

import "github.com/ElrondNetwork/elrond-go/crypto"

// RedundancyHandlerStub -
type RedundancyHandlerStub struct {
	IsRedundancyNodeCalled    func() bool
	IsMainMachineActiveCalled func() bool
	ObserverPrivateKeyCalled  func() crypto.PrivateKey
}

// IsRedundancyNode -
func (rhs *RedundancyHandlerStub) IsRedundancyNode() bool {
	if rhs.IsRedundancyNodeCalled != nil {
		return rhs.IsRedundancyNodeCalled()
	}

	return false
}

// IsMainMachineActive -
func (rhs *RedundancyHandlerStub) IsMainMachineActive() bool {
	if rhs.IsMainMachineActiveCalled != nil {
		return rhs.IsMainMachineActiveCalled()
	}

	return true
}

// ObserverPrivateKey -
func (rhs *RedundancyHandlerStub) ObserverPrivateKey() crypto.PrivateKey {
	if rhs.ObserverPrivateKeyCalled != nil {
		return rhs.ObserverPrivateKeyCalled()
	}

	return &PrivateKeyStub{}
}

// IsInterfaceNil -
func (rhs *RedundancyHandlerStub) IsInterfaceNil() bool {
	return rhs == nil
}
