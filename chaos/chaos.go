package chaos

import (
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("chaos")

var Controller *chaosController

func init() {
	Controller = newChaosController("chaos/config.json")
}
