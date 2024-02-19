package scrProcFactory

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/smartContract/processProxy"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// TODO: Marius C./Radu C./Octavian: When merging feat/chain-go-sdk-refactor into feat/chain-go-sdk, please delete
// this completely and use the created factories

// CreateSCRProcessor creates a scr processor based on the chain run type (normal/sovereign)
func CreateSCRProcessor(chainRunType common.ChainRunType, scProcArgs scrCommon.ArgsNewSmartContractProcessor, epochNotifier vmcommon.EpochNotifier) (scrCommon.SCRProcessorHandler, error) {
	switch chainRunType {
	case common.ChainRunTypeRegular:
		return processProxy.NewSmartContractProcessorProxy(scProcArgs, epochNotifier)
	case common.ChainRunTypeSovereign:
		scrProc, err := processorV2.NewSmartContractProcessorV2(scProcArgs)
		if err != nil {
			return nil, err
		}
		return processorV2.NewSovereignSCRProcessor(scrProc)
	default:
		return nil, fmt.Errorf("%w type %v", customErrors.ErrUnimplementedChainRunType, chainRunType)
	}
}
