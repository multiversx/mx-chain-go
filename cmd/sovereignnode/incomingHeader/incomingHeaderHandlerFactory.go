package incomingHeader

import (
	hasherFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	marshallerFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	mainFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
)

func CreateIncomingHeaderProcessor(
	config *config.NotifierConfig,
	dataPool dataRetriever.PoolsHolder,
	mainChainNotarizationStartRound uint64,
	runTypeComponents mainFactory.RunTypeComponentsHolder,
) (process.IncomingHeaderSubscriber, error) {
	marshaller, err := marshallerFactory.NewMarshalizer(config.WebSocketConfig.MarshallerType)
	if err != nil {
		return nil, err
	}
	hasher, err := hasherFactory.NewHasher(config.WebSocketConfig.HasherType)
	if err != nil {
		return nil, err
	}

	argsIncomingHeaderHandler := ArgsIncomingHeaderProcessor{
		HeadersPool:                     dataPool.Headers(),
		TxPool:                          dataPool.UnsignedTransactions(),
		Marshaller:                      marshaller,
		Hasher:                          hasher,
		MainChainNotarizationStartRound: mainChainNotarizationStartRound,
		OutGoingOperationsPool:          runTypeComponents.OutGoingOperationsPoolHandler(),
		DataCodec:                       runTypeComponents.DataCodecHandler(),
		TopicsChecker:                   runTypeComponents.TopicsCheckerHandler(),
	}

	return NewIncomingHeaderProcessor(argsIncomingHeaderHandler)
}
