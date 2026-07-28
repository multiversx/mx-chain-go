package components

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process"
)

const defaultCreateBlockMaxTimePercent = 0.25

type roundDurationHandler interface {
	TimeDuration() time.Duration
}

// createBlockTimeBoundProcessor keeps the simulator's create-block time limit active when blocks
// are proposed through SPoS. The direct simulator path applies the same limit in blocksCreator,
// which is bypassed entirely when consensus is enabled.
type createBlockTimeBoundProcessor struct {
	process.BlockProcessor

	roundHandler              roundDurationHandler
	createBlockMaxTimePercent float64
	now                       func() time.Time
}

func newCreateBlockTimeBoundProcessor(
	blockProcessor process.BlockProcessor,
	roundHandler roundDurationHandler,
	createBlockMaxTimePercent float64,
) *createBlockTimeBoundProcessor {
	if createBlockMaxTimePercent == 0 {
		createBlockMaxTimePercent = defaultCreateBlockMaxTimePercent
	}

	return &createBlockTimeBoundProcessor{
		BlockProcessor:            blockProcessor,
		roundHandler:              roundHandler,
		createBlockMaxTimePercent: createBlockMaxTimePercent,
		now:                       time.Now,
	}
}

func (processor *createBlockTimeBoundProcessor) CreateBlock(
	initialHeader data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	return processor.BlockProcessor.CreateBlock(initialHeader, processor.withTimeLimit(haveTime))
}

func (processor *createBlockTimeBoundProcessor) CreateBlockProposal(
	initialHeader data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	return processor.BlockProcessor.CreateBlockProposal(initialHeader, processor.withTimeLimit(haveTime))
}

func (processor *createBlockTimeBoundProcessor) withTimeLimit(haveTime func() bool) func() bool {
	startTime := processor.now()
	allowedDuration := time.Duration(
		float64(processor.roundHandler.TimeDuration()) * processor.createBlockMaxTimePercent,
	)

	return func() bool {
		return haveTime() && processor.now().Sub(startTime) < allowedDuration
	}
}
