package cutoff

import (
	"github.com/multiversx/mx-chain-core-go/data/endProcess"

	"github.com/multiversx/mx-chain-go/config"
)

// CreateBlockProcessingCutoffHandler will create the desired block processing cutoff handler based on configuration
func CreateBlockProcessingCutoffHandler(
	cfg config.BlockProcessingCutoffConfig,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (BlockProcessingCutoffHandler, error) {
	if !cfg.Enabled {
		return NewDisabledBlockProcessingCutoff(), nil
	}

	return NewBlockProcessingCutoffHandler(cfg, chanStopNodeProcess)
}
