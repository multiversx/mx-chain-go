package block

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/process"
)

// executionResultsVerifier is a struct that checks the execution results of a shard header
type executionResultsVerifier struct {
	blockChain              data.ChainHandler
	executionResultsTracker ExecutionResultsTracker
}

// NewExecutionResultsVerifier creates a new instance of executionResultsVerifier
func NewExecutionResultsVerifier(blockChain data.ChainHandler, executionResultsTracker ExecutionResultsTracker) (*executionResultsVerifier, error) {
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
func (erc *executionResultsVerifier) VerifyHeaderExecutionResults(header data.ShardHeaderHandler, headerHash []byte) error {
	if !header.IsHeaderV3() {
		return process.ErrInvalidHeader
	}
	if header.GetNonce() < 1 {
		return nil
	}

	executionResults := header.GetExecutionResultsHandlers()
	lastExecutionResultInfo := header.GetLastExecutionResultHandler()

	return erc.verifyExecutionResults(executionResults, headerHash, lastExecutionResultInfo)
}

func (erc *executionResultsVerifier) verifyExecutionResults(executionResults []data.ExecutionResultHandler, headerHash []byte, lastExecutionResultInfo data.ShardExecutionResultInfo) error {
	if check.IfNil(lastExecutionResultInfo) {
		return fmt.Errorf("%w: for current block", process.ErrNilLastExecutionResultHandler)
	}

	prevLastExecutionResultInfo, err := erc.getPrevBlockLastExecutionResult()
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

	pendingExecutionResults, err := erc.executionResultsTracker.GetPendingExecutionResults()
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

func (erc *executionResultsVerifier) getPrevBlockLastExecutionResult() (data.ShardExecutionResultInfo, error) {
	prevHeader := erc.blockChain.GetCurrentBlockHeader()
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

	return erc.createLastExecutionResultFromPrevHeader(prevShardHeader, prevHeaderHash)
}

func (erc *executionResultsVerifier) createLastExecutionResultFromPrevHeader(prevShardHeader data.ShardHeaderHandler, prevShardHeaderHash []byte) (data.ShardExecutionResultInfo, error) {
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

func createLastExecutionResultInfoFromExecutionResult(notarizedOnHeaderHash []byte, lastExecResult data.ExecutionResultHandler) (*block.ExecutionResultInfo, error) {
	if len(notarizedOnHeaderHash) == 0 {
		return nil, process.ErrNilNotarizedOnHeaderHash
	}
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
