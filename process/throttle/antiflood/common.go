package antiflood

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

// FloodPreventerConfigFetcher defines custom config handler func for flood preventer
type FloodPreventerConfigFetcher func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig

// OutputIdentifier defines output identifier for antiflodd components
const OutputIdentifier = "output"
