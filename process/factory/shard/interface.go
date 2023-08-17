package shard

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
)

// SmartContractResultPreProcessorCreator defines the interface of a smart contract result pre-processor creator
type SmartContractResultPreProcessorCreator interface {
	CreateSmartContractResultPreProcessor(args preprocess.SmartContractResultPreProcessorCreatorArgs) (process.PreProcessor, error)
	IsInterfaceNil() bool
}
