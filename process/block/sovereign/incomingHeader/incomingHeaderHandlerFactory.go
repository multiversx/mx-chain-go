package incomingHeader

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"

	hasherFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	marshallerFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
)

type RunTypeComponentsHolder interface {
	OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool
	DataCodecHandler() sovereign.DataCodecHandler
	TopicsCheckerHandler() sovereign.TopicsCheckerHandler
}

// CreateIncomingHeaderProcessor creates the incoming header processor
func CreateIncomingHeaderProcessor(
	config *config.NotifierConfig,
	dataPool dataRetriever.PoolsHolder,
	mainChainNotarizationStartRound uint64,
	runTypeComponents RunTypeComponentsHolder,
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
