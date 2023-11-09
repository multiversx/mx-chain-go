package mock

import "github.com/multiversx/mx-chain-go/update"

// ExportFactoryHandlerStub -
type ExportFactoryHandlerStub struct {
	CreateCalled func() (update.ExportHandler, error)
}

// Create -
func (e *ExportFactoryHandlerStub) Create() (update.ExportHandler, error) {
	if e.CreateCalled != nil {
		return e.CreateCalled()
	}
	return &ExportHandlerStub{}, nil
}

// IsInterfaceNil -
func (e *ExportFactoryHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
