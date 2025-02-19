package chaos

import (
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("chaos")
var Controller *chaosController

// Initialize initializes the chaos controller. Make sure to call this only after logging components (file logging, as well) are set up.
func Initialize() {
	log.Info("initializing chaos controller")
	Controller = newChaosController(defaultConfigFilePath)
}
