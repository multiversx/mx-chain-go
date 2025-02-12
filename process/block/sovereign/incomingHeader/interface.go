package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"

	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	sovBlock "github.com/multiversx/mx-chain-go/process/block/sovereign"
)

// HeadersPool should be able to add new headers in pool
type HeadersPool interface {
	AddHeaderInShard(headerHash []byte, header data.HeaderHandler, shardID uint32)
	IsInterfaceNil() bool
}

// TransactionPool should be able to add a new transaction in the pool
type TransactionPool interface {
	AddData(key []byte, data interface{}, sizeInBytes int, cacheId string)
	IsInterfaceNil() bool
}

// TopicsChecker should be able to check the topics validity
type TopicsChecker interface {
	CheckValidity(topics [][]byte) error
	IsInterfaceNil() bool
}

// SovereignDataCodec is the interface for serializing/deserializing data
type SovereignDataCodec interface {
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	IsInterfaceNil() bool
}

// RunTypeComponentsHolder defines run type components needed to create an incoming header processor
type RunTypeComponentsHolder interface {
	OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool
	DataCodecHandler() sovBlock.DataCodecHandler
	TopicsCheckerHandler() sovBlock.TopicsCheckerHandler
	IsInterfaceNil() bool
}

// IncomingEventHandler defines the behaviour of an incoming cross chain event processor handler
type IncomingEventHandler interface {
	ProcessEvent(event data.EventHandler) (*EventResult, error)
	IsInterfaceNil() bool
}
