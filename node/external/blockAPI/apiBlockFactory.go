package blockAPI

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// CreateAPIBlockProcessor will create a new instance of APIBlockHandler
func CreateAPIBlockProcessor(arg *ArgAPIBlockProcessor) (APIBlockHandler, error) {
	err := checkNilArg(arg)
	if err != nil {
		return nil, err
	}

	emptyReceiptsHash, err := computeEmptyReceiptsHash(arg.Marshalizer, arg.Hasher)
	if err != nil {
		return nil, err
	}

	if arg.SelfShardID != core.MetachainShardId {
		return newShardApiBlockProcessor(arg, emptyReceiptsHash), nil
	}

	return newMetaApiBlockProcessor(arg, emptyReceiptsHash), nil
}

// CreateAPIInternalBlockProcessor will create a new instance of APIInternalBlockHandler
func CreateAPIInternalBlockProcessor(arg *ArgAPIBlockProcessor) (APIInternalBlockHandler, error) {
	err := checkNilArg(arg)
	if err != nil {
		return nil, err
	}

	emptyReceiptsHash, err := computeEmptyReceiptsHash(arg.Marshalizer, arg.Hasher)
	if err != nil {
		return nil, err
	}

	return newInternalBlockProcessor(arg, emptyReceiptsHash), nil
}

func computeEmptyReceiptsHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	allReceiptsHashes := make([][]byte, 0)

	return core.CalculateHash(marshalizer, hasher, &batch.Batch{Data: allReceiptsHashes})
}
