package frozen

import (
	"os"
	"strconv"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var IsFrozen = false
var ShouldOverrideHighestRound = false
var HighestRound = int64(0)

var log = logger.GetOrCreate("frozen")

func Setup() {
	envFrozen := os.Getenv("FROZEN")
	envHighestRound := os.Getenv("HIGHEST_ROUND")

	IsFrozen, _ = strconv.ParseBool(envFrozen)
	ShouldOverrideHighestRound = len(envHighestRound) > 0
	HighestRound, _ = strconv.ParseInt(envHighestRound, 10, 64)

	if IsFrozen {
		log.Warn("Frozen mode is enabled")
	}

	if ShouldOverrideHighestRound {
		log.Warn("Highest round is overriden", "round", HighestRound)
	}
}
