package factory

import (
	mxFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/update"
	updateMock "github.com/multiversx/mx-chain-go/update/mock"
)

// ExportHandlerFactoryCreatorMock -
type ExportHandlerFactoryCreatorMock struct {
	CreateExportFactoryHandlerCalled func(args mxFactory.ArgsExporter) (update.ExportFactoryHandler, error)
}

// CreateExportFactoryHandler -
func (mock *ExportHandlerFactoryCreatorMock) CreateExportFactoryHandler(args mxFactory.ArgsExporter) (update.ExportFactoryHandler, error) {
	if mock.CreateExportFactoryHandlerCalled != nil {
		return mock.CreateExportFactoryHandlerCalled(args)
	}

	return &updateMock.ExportFactoryHandlerStub{}, nil
}

// IsInterfaceNil -
func (mock *ExportHandlerFactoryCreatorMock) IsInterfaceNil() bool {
	return mock == nil
}
