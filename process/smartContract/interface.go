package smartContract

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
)

// SCProcessorCreator defines a scr processor creator
type SCProcessorCreator interface {
	CreateSCProcessor(args scrCommon.ArgsNewSmartContractProcessor, epochNotifier process.EpochNotifier) (scrCommon.SCRProcessorHandler, error)
	IsInterfaceNil() bool
}
