package block

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

// CreateNewHeaderProposal creates a new header
func (mp *metaProcessor) CreateNewHeaderProposal(round uint64, nonce uint64) (data.HeaderHandler, error) {
	epoch := mp.epochStartTrigger.Epoch()

	header := mp.versionedHeaderFactory.Create(epoch, round)
	metaHeader, ok := header.(data.MetaHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	if !metaHeader.IsHeaderV3() {
		return nil, process.ErrInvalidHeader
	}

	epochChangeProposed := mp.epochStartTrigger.ShouldProposeEpochChange(round, nonce)
	metaHeader.SetEpochChangeProposed(epochChangeProposed)
	err := metaHeader.SetRound(round)
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetNonce(nonce)
	if err != nil {
		return nil, err
	}

	err = mp.addExecutionResultsOnHeader(metaHeader)
	if err != nil {
		return nil, err
	}

	hasEpochStartResults, err := mp.hasStartOfEpochExecutionResults(metaHeader)
	if err != nil {
		return nil, err
	}
	if hasEpochStartResults {
		err := metaHeader.SetEpoch(epoch + 1)
		if err != nil {
			return nil, err
		}
	}

	// TODO: the trigger would need to be changed upon commit of a block with the epoch start results

	return metaHeader, nil
}

// CreateBlockProposal creates a block proposal without executing any of the transactions
func (mp *metaProcessor) CreateBlockProposal(
	initialHdr data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}
	if !initialHdr.IsHeaderV3() {
		return nil, nil, process.ErrInvalidHeader
	}

	metaHdr, ok := initialHdr.(*block.MetaBlockV3)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	metaHdr.SoftwareVersion = []byte(mp.headerIntegrityVerifier.GetVersion(metaHdr.Epoch, metaHdr.Round))

	// var body data.BodyHandler
	// err := mp.updateHeaderForEpochStartIfNeeded(metaHdr)
	// if err != nil {
	// 	return nil, nil, err
	// }
	//
	// err = mp.blockChainHook.SetCurrentHeader(metaHdr)
	// if err != nil {
	// 	return nil, nil, err
	// }
	//
	// body, err = mp.createBody(metaHdr, haveTime)
	// if err != nil {
	// 	return nil, nil, err
	// }
	//
	// body, err = mp.applyBodyToHeader(metaHdr, body)
	// if err != nil {
	// 	return nil, nil, err
	// }
	//
	// mp.requestHandler.SetEpoch(metaHdr.GetEpoch())
	// return metaHdr, body, nil

	return nil, nil, nil
}

// VerifyBlockProposal will be implemented in a further PR
func (mp *metaProcessor) VerifyBlockProposal(
	_ data.HeaderHandler,
	_ data.BodyHandler,
	_ func() time.Duration,
) error {
	return nil
}

// ProcessBlockProposal processes the proposed block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlockProposal(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (data.BaseExecutionResultHandler, error) {
	return nil, nil
}

func (mp *metaProcessor) hasStartOfEpochExecutionResults(metaHeader data.MetaHeaderHandler) (bool, error) {
	execResults := metaHeader.GetExecutionResultsHandlers()
	for _, execResult := range execResults {
		mbHeaders, err := mp.extractMiniBlocksHeaderHandlersFromExecResult(execResult, common.MetachainShardId)
		if err != nil {
			return false, err
		}
		if hasRewardMiniBlocks(mbHeaders) {
			return true, nil
		}
	}
	return false, nil
}

func hasRewardMiniBlocks(miniBlockHeaders []data.MiniBlockHeaderHandler) bool {
	for _, mbHeader := range miniBlockHeaders {
		if mbHeader.GetTypeInt32() == int32(block.RewardsBlock) ||
			mbHeader.GetTypeInt32() == int32(block.PeerBlock) {
			return true
		}
	}
	return false
}
