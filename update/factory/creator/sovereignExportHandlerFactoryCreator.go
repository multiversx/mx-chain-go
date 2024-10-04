package creator

import (
	mxFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/disabled"
)

type sovereignExportHandlerFactoryCreator struct {
}

// NewSovereignExportHandlerFactoryCreator creates a sovereign export handler factory creator
func NewSovereignExportHandlerFactoryCreator() *sovereignExportHandlerFactoryCreator {
	return &sovereignExportHandlerFactoryCreator{}
}

// CreateExportFactoryHandler creates a disabled export factory handler
func (f *sovereignExportHandlerFactoryCreator) CreateExportFactoryHandler(_ mxFactory.ArgsExporter) (update.ExportFactoryHandler, error) {
	return &disabled.ExportFactoryHandler{}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignExportHandlerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
