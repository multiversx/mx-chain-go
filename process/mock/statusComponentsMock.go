package mock

import (
	"github.com/multiversx/mx-chain-go/outport"
)

// StatusComponentsMock -
type StatusComponentsMock struct {
	Outport outport.OutportHandler
}

// OutportHandler -
func (scm *StatusComponentsMock) OutportHandler() outport.OutportHandler {
	return scm.Outport
}

// IsInterfaceNil -
func (scm *StatusComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
