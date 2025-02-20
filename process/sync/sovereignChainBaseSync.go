package sync

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
)

func (boot *baseBootstrap) sovereignChainProcessAndCommit(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
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

func (boot *baseBootstrap) sovereignChainHandleScheduledRollBackToHeader(header data.HeaderHandler, headerHash []byte) []byte {
	return boot.sovereignChainGetRootHashFromBlock(header, headerHash)
}

func (boot *baseBootstrap) sovereignChainGetRootHashFromBlock(header data.HeaderHandler, _ []byte) []byte {
	rootHash := boot.chainHandler.GetGenesisHeader().GetRootHash()
	if header != nil {
		rootHash = header.GetRootHash()
	}

	return rootHash
}

func (boot *baseBootstrap) sovereignChainDoProcessReceivedHeaderJob(headerHandler data.HeaderHandler, headerHash []byte) {
	_, isExtendedShardHeaderReceived := headerHandler.(*block.ShardHeaderExtended)
	if isExtendedShardHeaderReceived {
		return
	}

	boot.doProcessReceivedHeaderJob(headerHandler, headerHash)
}
