package mock

import (
	"github.com/ElrondNetwork/elrond-go/outport"
)

// StatusComponentsMock -
type StatusComponentsMock struct {
	Outport      outport.OutportHandler
	TPSBenchmark statistics.TPSBenchmark
}

// OutportHandler -
func (scm *StatusComponentsMock) OutportHandler() outport.OutportHandler {
	return scm.Outport
}

// IsInterfaceNil -
func (scm *StatusComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
