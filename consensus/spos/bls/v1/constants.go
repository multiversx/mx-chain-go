package v1

import (
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("consensus/spos/bls/v1")

// waitingAllSigsMaxTimeThreshold specifies the max allocated time for waiting all signatures from the total time of the subround signature
const waitingAllSigsMaxTimeThreshold = 0.5
