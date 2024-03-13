//go:build !arm64

package node

import (
	"runtime"

	"github.com/multiversx/mx-chain-go/config"
)

func applyArchCustomConfigs(_ *config.Configs) {
	log.Debug("applyArchCustomConfigs - nothing to do", "architecture", runtime.GOARCH)
}
