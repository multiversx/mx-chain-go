package shard

import (
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
)

// RunTypeComponents defines the needed run type components for pre-processor factory
type RunTypeComponents interface {
	TxPreProcessorCreator() preprocess.TxPreProcessorCreator
	SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator
	IsInterfaceNil() bool
}
