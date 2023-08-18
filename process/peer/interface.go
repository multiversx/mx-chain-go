package peer

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// DataPool indicates the main functionality needed in order to fetch the required blocks from the pool
type DataPool interface {
	Headers() dataRetriever.HeadersPool
	IsInterfaceNil() bool
}

// ValidatorStatisticsProcessorCreator is an interface for creating validator statistics processors
type ValidatorStatisticsProcessorCreator interface {
	CreateValidatorStatisticsProcessor(args ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error)
	IsInterfaceNil() bool
}
