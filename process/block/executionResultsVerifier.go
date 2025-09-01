package block

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/process"
)

// executionResultsVerifier is a struct that checks the execution results of a shard header
type executionResultsVerifier struct {
	blockChain              data.ChainHandler
	executionResultsTracker process.ExecutionResultsTracker
}

// NewExecutionResultsVerifier creates a new instance of executionResultsVerifier
func NewExecutionResultsVerifier(blockChain data.ChainHandler, executionResultsTracker process.ExecutionResultsTracker) (*executionResultsVerifier, error) {
	if check.IfNil(blockChain) {
		return nil, process.ErrNilBlockChain
	}
	if check.IfNil(executionResultsTracker) {
		return nil, process.ErrNilExecutionResultsTracker
	}
	return &executionResultsVerifier{
		blockChain:              blockChain,
		executionResultsTracker: executionResultsTracker,
	}, nil
}

// VerifyHeaderExecutionResults checks the execution results of a shard header
func (erc *executionResultsVerifier) VerifyHeaderExecutionResults(headerHash []byte, header data.HeaderHandler) error {
	if len(headerHash) == 0 {
		return process.ErrInvalidHash
	}
	if check.IfNil(header) {
		return process.ErrNilHeaderHandler
	}
	if !header.IsHeaderV3() {
		return process.ErrInvalidHeader
	}
	if header.GetNonce() < 1 {
		return nil
	}

	return erc.verifyExecutionResults(headerHash, header)
}

func (erc *executionResultsVerifier) verifyExecutionResults(
	headerHash []byte,
	header data.HeaderHandler,
) error {
	err := erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(headerHash, header)
	if err != nil {
		return err
	}

	// if header is received, the notarized execution results should already be available in the tracker
	// if not all present, then verify fails
	executionResults := header.GetExecutionResultsHandlers()
	pendingExecutionResults, err := erc.executionResultsTracker.GetPendingExecutionResults()
	if err != nil {
		return err
	}

	if len(pendingExecutionResults) < len(executionResults) {
		return process.ErrExecutionResultsNumberMismatch
	}

	for i := 0; i < len(executionResults)-1; i++ {
		if executionResults[i].GetHeaderNonce() != executionResults[i+1].GetHeaderNonce()-1 {
			return process.ErrExecutionResultsNonConsecutive
		}
	}

	for i, er := range executionResults {
		if !er.Equal(pendingExecutionResults[i]) {
			return process.ErrExecutionResultDoesNotMatch
		}
	}

	return nil
}

func (erc *executionResultsVerifier) verifyLastExecutionResultInfoMatchesLastExecutionResult(
	headerHash []byte,
	header data.HeaderHandler,
) error {
	executionResults := header.GetExecutionResultsHandlers()
	lastExecutionResultInfo := header.GetLastExecutionResultHandler()
	shardID := header.GetShardID()
	if check.IfNil(lastExecutionResultInfo) {
		return fmt.Errorf("%w: for current block", process.ErrNilLastExecutionResultHandler)
	}

	prevLastExecutionResultInfo, err := erc.getPrevBlockLastExecutionResult()
	if err != nil {
		return err
	}

	// if no execution results are present, we only check if the last execution result info matches the previous reported one
	if len(executionResults) == 0 {
		if !lastExecutionResultInfo.Equal(prevLastExecutionResultInfo) {
			return process.ErrExecutionResultDoesNotMatch
		}

		return nil
	}

	lastExecResult := executionResults[len(executionResults)-1]
	lastExecResultInfo, err := createLastExecutionResultInfoFromExecutionResult(headerHash, lastExecResult, shardID)
	if err != nil {
		return err
	}

	if !lastExecutionResultInfo.Equal(lastExecResultInfo) {
		return process.ErrExecutionResultDoesNotMatch
	}

	err = erc.checkFirstExecutionResultAgainstPrevBlock(prevLastExecutionResultInfo, executionResults)
	if err != nil {
		return err
	}

	return nil
}

func (erc *executionResultsVerifier) checkFirstExecutionResultAgainstPrevBlock(
	prevLastExecutionResultsHandler data.LastExecutionResultHandler,
	executionResults []data.BaseExecutionResultHandler,
) error {
	var prevNonce uint64
	switch prev := prevLastExecutionResultsHandler.(type) {
	case *block.ExecutionResultInfo:
		prevNonce = prev.GetExecutionResult().GetHeaderNonce()
	case *block.MetaExecutionResultInfo:
		prevNonce = prev.GetExecutionResult().GetHeaderNonce()
	default:
		return process.ErrWrongTypeAssertion
	}
	// if execution results are present, we check if the previous last execution result info matches the first execution result
	if executionResults[0].GetHeaderNonce() != prevNonce+1 {
		return fmt.Errorf("%w for first execution result", process.ErrExecutionResultsNonConsecutive)
	}

	return nil
}

func (erc *executionResultsVerifier) getPrevBlockLastExecutionResult() (data.LastExecutionResultHandler, error) {
	prevHeader := erc.blockChain.GetCurrentBlockHeader()
	if check.IfNil(prevHeader) {
		return nil, process.ErrNilHeaderHandler
	}

	if prevHeader.IsHeaderV3() {
		return prevHeader.GetLastExecutionResultHandler(), nil
	}

	prevHeaderHash := erc.blockChain.GetCurrentBlockHeaderHash()

	return createLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
}

func createLastExecutionResultFromPrevHeader(prevHeader data.HeaderHandler, prevHeaderHash []byte) (data.LastExecutionResultHandler, error) {
	if check.IfNil(prevHeader) {
		return nil, process.ErrNilBlockHeader
	}
	if len(prevHeaderHash) == 0 {
		return nil, process.ErrInvalidHash
	}

	if prevHeader.GetShardID() != core.MetachainShardId {
		if _, ok := prevHeader.(*block.HeaderV2); !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		return &block.ExecutionResultInfo{
			NotarizedOnHeaderHash: prevHeaderHash,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  prevHeaderHash,
				HeaderNonce: prevHeader.GetNonce(),
				HeaderRound: prevHeader.GetRound(),
				RootHash:    prevHeader.GetRootHash(),
			},
		}, nil
	}

	prevMetaHeader, ok := prevHeader.(*block.MetaBlock)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return &block.MetaExecutionResultInfo{
		NotarizedOnHeaderHash: prevHeaderHash,
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  prevHeaderHash,
				HeaderNonce: prevMetaHeader.GetNonce(),
				HeaderRound: prevMetaHeader.GetRound(),
				RootHash:    prevMetaHeader.GetRootHash(),
			},
			ValidatorStatsRootHash: prevMetaHeader.GetValidatorStatsRootHash(),
			AccumulatedFeesInEpoch: prevMetaHeader.GetAccumulatedFeesInEpoch(),
			DevFeesInEpoch:         prevMetaHeader.GetDevFeesInEpoch(),
		},
	}, nil
}

func createLastExecutionResultInfoFromExecutionResult(notarizedOnHeaderHash []byte, lastExecResult data.BaseExecutionResultHandler, shardID uint32) (data.LastExecutionResultHandler, error) {
	if len(notarizedOnHeaderHash) == 0 {
		return nil, process.ErrNilNotarizedOnHeaderHash
	}
	if check.IfNil(lastExecResult) {
		return nil, process.ErrNilExecutionResultHandler
	}

	if shardID != core.MetachainShardId {
		if _, ok := lastExecResult.(*block.ExecutionResult); !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		return &block.ExecutionResultInfo{
			NotarizedOnHeaderHash: notarizedOnHeaderHash,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  lastExecResult.GetHeaderHash(),
				HeaderNonce: lastExecResult.GetHeaderNonce(),
				HeaderRound: lastExecResult.GetHeaderRound(),
				RootHash:    lastExecResult.GetRootHash(),
			},
		}, nil
	}

	lastMetaExecResult, ok := lastExecResult.(*block.MetaExecutionResult)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return &block.MetaExecutionResultInfo{
		NotarizedOnHeaderHash: notarizedOnHeaderHash,
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  lastMetaExecResult.GetHeaderHash(),
				HeaderNonce: lastMetaExecResult.GetHeaderNonce(),
				HeaderRound: lastMetaExecResult.GetHeaderRound(),
				RootHash:    lastMetaExecResult.GetRootHash(),
			},
			ValidatorStatsRootHash: lastMetaExecResult.GetValidatorStatsRootHash(),
			AccumulatedFeesInEpoch: lastMetaExecResult.GetAccumulatedFeesInEpoch(),
			DevFeesInEpoch:         lastMetaExecResult.GetDevFeesInEpoch(),
		},
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (erc *executionResultsVerifier) IsInterfaceNil() bool {
	return erc == nil
}
