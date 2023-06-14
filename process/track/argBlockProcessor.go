package track

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgBlockProcessor holds all dependencies required to process tracked blocks in order to create new instances of
// block processor
type ArgBlockProcessor struct {
	HeaderValidator                       process.HeaderConstructionValidator
	RequestHandler                        process.RequestHandler
	ShardCoordinator                      sharding.Coordinator
	ChainParametersHandler                process.ChainParametersHandler
	InitialBlockFinality                  uint64
	BlockTracker                          blockTrackerHandler
	CrossNotarizer                        blockNotarizerHandler
	SelfNotarizer                         blockNotarizerHandler
	CrossNotarizedHeadersNotifier         blockNotifierHandler
	SelfNotarizedFromCrossHeadersNotifier blockNotifierHandler
	SelfNotarizedHeadersNotifier          blockNotifierHandler
	FinalMetachainHeadersNotifier         blockNotifierHandler
	RoundHandler                          process.RoundHandler
}
