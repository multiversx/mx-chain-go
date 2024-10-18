package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	hasherFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	marshallerFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// CreateIncomingHeaderProcessor creates the incoming header processor
func CreateIncomingHeaderProcessor(
	config config.WebSocketConfig,
	dataPool dataRetriever.PoolsHolder,
	mainChainNotarizationStartRound uint64,
	runTypeComponents RunTypeComponentsHolder,
) (process.IncomingHeaderSubscriber, error) {
	if check.IfNil(runTypeComponents) {
		return nil, errorsMx.ErrNilRunTypeComponents
	}
	marshaller, err := marshallerFactory.NewMarshalizer(config.MarshallerType)
	if err != nil {
		return nil, err
	}
	hasher, err := hasherFactory.NewHasher(config.HasherType)
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
