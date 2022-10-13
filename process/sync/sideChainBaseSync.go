package sync

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
)

func (boot *baseBootstrap) sideChainProcessAndCommit(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	startProcessBlockTime := time.Now()
	header, body, err := boot.blockProcessor.ProcessBlock(header, body, haveTime)
	elapsedTime := time.Since(startProcessBlockTime)
	log.Debug("elapsed time to process block",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	startCommitBlockTime := time.Now()
	err = boot.blockProcessor.CommitBlock(header, body)
	elapsedTime = time.Since(startCommitBlockTime)
	if elapsedTime >= common.CommitMaxTime {
		log.Warn("syncBlock.CommitBlock", "elapsed time", elapsedTime)
	} else {
		log.Debug("elapsed time to commit block",
			"time [s]", elapsedTime,
		)
	}

	return err
}
