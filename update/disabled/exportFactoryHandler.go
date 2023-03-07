package disabled

import "github.com/multiversx/mx-chain-go/update"

// ExportFactoryHandler implements ExportFactoryHandler interface but does nothing
type ExportFactoryHandler struct {
}

// Create does nothing as it is disabled
func (e *ExportFactoryHandler) Create() (update.ExportHandler, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *ExportFactoryHandler) IsInterfaceNil() bool {
	return e == nil
}
