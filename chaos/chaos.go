package chaos

import (
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("chaos")
var Controller = newChaosController()
