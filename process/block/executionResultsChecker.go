package block

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/process"
)

func (sp *shardProcessor) checkExecutionResultsForHeader(header data.ShardHeaderHandler, headerHash []byte) error {
	if !header.IsHeaderV3() {
		return process.ErrInvalidHeader
	}
	if header.GetNonce() < 1 {
		return nil
	}

	executionResults := header.GetExecutionResultsHandlers()
	lastExecutionResultInfo := header.GetLastExecutionResultHandler()

	return sp.checkExecutionResults(executionResults, headerHash, lastExecutionResultInfo)
}

func (sp *shardProcessor) checkExecutionResults(executionResults []data.ExecutionResultHandler, headerHash []byte, lastExecutionResultInfo data.ShardExecutionResultInfo) error {
	if check.IfNil(lastExecutionResultInfo) {
		return fmt.Errorf("%w: for current block", process.ErrNilLastExecutionResultHandler)
	}

	prevLastExecutionResultInfo, err := sp.getPrevBlockLastExecutionResult()
	if err != nil {
		return err
	}

	if check.IfNil(prevLastExecutionResultInfo) {
		return fmt.Errorf("%w: for previous block", process.ErrNilLastExecutionResultHandler)
	}

	if len(executionResults) == 0 {
		if !lastExecutionResultInfo.Equal(prevLastExecutionResultInfo) {
			return process.ErrExecutionResultDoesNotMatch
		}

		return nil
	}

	lastExecResult := executionResults[len(executionResults)-1]
	lastExecResultInfo, err := createLastExecutionResultInfoFromExecutionResult(headerHash, lastExecResult)
	if err != nil {
		return err
	}

	if !lastExecutionResultInfo.Equal(lastExecResultInfo) {
		return process.ErrExecutionResultDoesNotMatch
	}

	// TODO: take this from the sp.executionResultsTracker
	var executionResultsTracker ExecutionResultsTracker

	pendingExecutionResults, err := executionResultsTracker.GetPendingExecutionResults()
	if err != nil {
		return err
	}

	if len(pendingExecutionResults) < len(executionResults) {
		return process.ErrExecutionResultsNumberMismatch
	}

	for i, er := range executionResults {
		if !er.Equal(pendingExecutionResults[i]) {
			return process.ErrExecutionResultMismatch
		}
	}

	return nil
}

func createLastExecutionResultInfoFromExecutionResult(notarizedOnHeaderHash []byte, lastExecResult data.ExecutionResultHandler) (*block.ExecutionResultInfo, error) {
	if check.IfNil(lastExecResult) {
		return nil, process.ErrNilExecutionResultHandler
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

func (sp *shardProcessor) getPrevBlockLastExecutionResult() (data.ShardExecutionResultInfo, error) {
	prevHeader := sp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(prevHeader) {
		return nil, process.ErrNilHeaderHandler
	}

	prevShardHeader, ok := prevHeader.(data.ShardHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	if prevShardHeader.IsHeaderV3() {
		return prevShardHeader.GetLastExecutionResultHandler(), nil
	}

	prevHeaderHash := prevHeader.GetPrevHash()

	return sp.createLastExecutionResultFromPrevHeader(prevShardHeader, prevHeaderHash)
}

func (sp *shardProcessor) createLastExecutionResultFromPrevHeader(prevShardHeader data.ShardHeaderHandler, prevShardHeaderHash []byte) (data.ShardExecutionResultInfo, error) {
	if prevShardHeader.IsHeaderV3() {
		return nil, process.ErrInvalidHeader
	}
	// create a dummy execution result info to be used in the case of Supernova
	execResult := &block.ExecutionResultInfo{
		NotarizedOnHeaderHash: prevShardHeaderHash,
		ExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  prevShardHeaderHash,
			HeaderNonce: prevShardHeader.GetNonce(),
			HeaderRound: prevShardHeader.GetRound(),
			RootHash:    prevShardHeader.GetRootHash(),
		},
	}

	return execResult, nil
}
