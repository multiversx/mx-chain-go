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
	blockChain       data.ChainHandler
	executionManager process.ExecutionManager
}

// NewExecutionResultsVerifier creates a new instance of executionResultsVerifier
func NewExecutionResultsVerifier(blockChain data.ChainHandler, executionManager process.ExecutionManager) (*executionResultsVerifier, error) {
	if check.IfNil(blockChain) {
		return nil, process.ErrNilBlockChain
	}
	if check.IfNil(executionManager) {
		return nil, process.ErrNilExecutionManager
	}
	return &executionResultsVerifier{
		blockChain:       blockChain,
		executionManager: executionManager,
	}, nil
}

// VerifyHeaderExecutionResults checks the execution results of a shard header
func (erc *executionResultsVerifier) VerifyHeaderExecutionResults(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return process.ErrNilHeaderHandler
	}
	if !header.IsHeaderV3() {
		return process.ErrInvalidHeader
	}
	if header.GetNonce() < 1 {
		return nil
	}

	return erc.verifyExecutionResults(header)
}

func (erc *executionResultsVerifier) verifyExecutionResults(
	header data.HeaderHandler,
) error {
	err := erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
	if err != nil {
		return err
	}

	// if header is received, the notarized execution results should already be available in the tracker
	// if not all present, then verify fails
	executionResults := header.GetExecutionResultsHandlers()
	pendingExecutionResults, err := erc.executionManager.GetPendingExecutionResults()
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
	header data.HeaderHandler,
) error {
	executionResults := header.GetExecutionResultsHandlers()
	lastExecutionResultInfo := header.GetLastExecutionResultHandler()
	shardID := header.GetShardID()
	if check.IfNil(lastExecutionResultInfo) {
		return fmt.Errorf("%w: for current block", process.ErrNilLastExecutionResultHandler)
	}

	prevLastExecutionResultInfo, err := process.GetPrevBlockLastExecutionResult(erc.blockChain)
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
	lastExecResultInfo, err := process.CreateLastExecutionResultInfoFromExecutionResult(header.GetRound(), lastExecResult, shardID)
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

// IsInterfaceNil returns true if there is no value under the interface
func (erc *executionResultsVerifier) IsInterfaceNil() bool {
	return erc == nil
}
