package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/data"

	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	sovBlock "github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/incomingEventsProc"
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

// RunTypeComponentsHolder defines run type components needed to create an incoming header processor
type RunTypeComponentsHolder interface {
	OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool
	DataCodecHandler() sovBlock.DataCodecHandler
	TopicsCheckerHandler() sovBlock.TopicsCheckerHandler
	IsInterfaceNil() bool
}

type IncomingEventsProcessor interface {
	RegisterProcessor(event string, proc incomingEventsProc.IncomingEventHandler) error
	ProcessIncomingEvents(events []data.EventHandler) (*dto.EventsResult, error)
}
