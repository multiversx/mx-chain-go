package api

import "github.com/multiversx/mx-chain-go/api/shared"

// UpgradeableHttpServerHandlerStub -
type UpgradeableHttpServerHandlerStub struct {
	StartHttpServerCalled func() error
	UpdateFacadeCalled    func(facade shared.FacadeHandler) error
	CloseCalled           func() error
}

// StartHttpServer -
func (stub *UpgradeableHttpServerHandlerStub) StartHttpServer() error {
	if stub.StartHttpServerCalled != nil {
		return stub.StartHttpServerCalled()
	}

	return nil
}

// UpdateFacade -
func (stub *UpgradeableHttpServerHandlerStub) UpdateFacade(facade shared.FacadeHandler) error {
	if stub.UpdateFacadeCalled != nil {
		return stub.UpdateFacadeCalled(facade)
	}

	return nil
}

// Close -
func (stub *UpgradeableHttpServerHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *UpgradeableHttpServerHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
