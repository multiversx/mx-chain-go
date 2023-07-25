package processorV2

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
)

// CreateSCRProcessor creates a scr processor based on the chain run type (normal/sovereign)
func CreateSCRProcessor(chainRunType common.ChainRunType, scProcArgs scrCommon.ArgsNewSmartContractProcessor) (scrCommon.SCRProcessorHandler, error) {
	scrProc, err := NewSmartContractProcessorV2(scProcArgs)

	switch chainRunType {
	case common.ChainRunTypeRegular:
		return scrProc, err
	case common.ChainRunTypeSovereign:
		return NewSovereignSCRProcessor(scrProc)
	default:
		return nil, fmt.Errorf("%w type %v", customErrors.ErrUnimplementedChainRunType, chainRunType)
	}
}
