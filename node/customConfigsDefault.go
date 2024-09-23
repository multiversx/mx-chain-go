//go:build !arm64

package node

import (
	"runtime"

	"github.com/multiversx/mx-chain-go/config"
)

// ApplyArchCustomConfigs will apply configuration tweaks based on the architecture the node is running on
func ApplyArchCustomConfigs(_ *config.Configs) {
	log.Debug("ApplyArchCustomConfigs - nothing to do", "architecture", runtime.GOARCH)
}
