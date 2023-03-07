package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug/process"
)

// CreateProcessDebugger creates a new instance of type ProcessDebugger
func CreateProcessDebugger(configs config.ProcessDebugConfig) (ProcessDebugger, error) {
	if !configs.Enabled {
		return process.NewDisabledDebugger(), nil
	}

	return process.NewProcessDebugger(configs)
}
