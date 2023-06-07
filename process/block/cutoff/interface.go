package cutoff

import "github.com/multiversx/mx-chain-core-go/data"

// BlockProcessingCutoffHandler defines the actions that a block processing handler has to take care of
type BlockProcessingCutoffHandler interface {
	HandleProcessErrorCutoff(header data.HeaderHandler) error
	HandlePauseCutoff(header data.HeaderHandler)
	IsInterfaceNil() bool
}
