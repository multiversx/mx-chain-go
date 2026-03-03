package estimator

import (
	"crypto/rand"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
)

type execResSizeComputation struct {
	execResSize    uint32
	maxExecResSize uint32
}

// NewExecResultSizeComputationHandler creates a execResSizeComputation instance
func NewExecResultSizeComputationHandler(
	marshalizer marshal.Marshalizer,
	maxExecResSize uint32,
) (*execResSizeComputation, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	ers := &execResSizeComputation{
		maxExecResSize: maxExecResSize,
	}

	var err error
	ers.execResSize, err = ers.generateDummyExecutionResultSize(marshalizer, 10)
	if err != nil {
		return nil, err
	}

	return ers, nil
}

func (ers *execResSizeComputation) generateDummyExecutionResultSize(
	marshaller marshal.Marshalizer,
	numMbs int,
) (uint32, error) {
	dummyHash := make([]byte, 32)
	_, _ = rand.Reader.Read(dummyHash)

	bigIntValue, _ := big.NewInt(0).SetString("10000000000000000000", 10)

	executionResult := &block.MetaExecutionResult{
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  dummyHash,
				HeaderNonce: 1,
				HeaderRound: 2,
				HeaderEpoch: 3,
				RootHash:    dummyHash,
				GasUsed:     1234,
			},
			ValidatorStatsRootHash: dummyHash,
			AccumulatedFeesInEpoch: bigIntValue,
			DevFeesInEpoch:         bigIntValue,
		},
		ReceiptsHash:    dummyHash,
		ExecutedTxCount: 10,
		AccumulatedFees: bigIntValue,
		DeveloperFees:   bigIntValue,
	}

	executionResult.MiniBlockHeaders = make([]block.MiniBlockHeader, numMbs)
	for i := 0; i < numMbs; i++ {
		executionResult.MiniBlockHeaders[i] = block.MiniBlockHeader{
			Hash:            dummyHash,
			SenderShardID:   1,
			ReceiverShardID: 2,
			TxCount:         10,
			Type:            1,
		}
	}

	buff, err := marshaller.Marshal(executionResult)
	if err != nil {
		return 0, err
	}

	return uint32(len(buff)), nil
}

// NewComputation will create a new size limit checker based on the precalculated execution result size
func (ers *execResSizeComputation) NewComputation() ExecResSizeLimitCheckerHandler {
	return newExecResultSizeLimitChecker(ers.execResSize, ers.maxExecResSize)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ers *execResSizeComputation) IsInterfaceNil() bool {
	return ers == nil
}
