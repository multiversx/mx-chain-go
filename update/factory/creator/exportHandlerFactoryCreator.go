package creator

import (
	mxFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/factory"
)

type exportHandlerFactoryCreator struct {
}

// NewExportHandlerFactoryCreator creates an export handler factory creator
func NewExportHandlerFactoryCreator() *exportHandlerFactoryCreator {
	return &exportHandlerFactoryCreator{}
}

// CreateExportFactoryHandler creates an export factory handler
func (f *exportHandlerFactoryCreator) CreateExportFactoryHandler(args mxFactory.ArgsExporter) (update.ExportFactoryHandler, error) {
	return factory.NewExportHandlerFactory(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *exportHandlerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
