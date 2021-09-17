package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ScrProcessor implements the SmartContractResultProcessor interface but does nothing as it is disabled
type ScrProcessor struct {
}

// ProcessSmartContractResult does nothing as it is disabled
func (scrProc *ScrProcessor) ProcessSmartContractResult(_ *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	return 0, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scrProc *ScrProcessor) IsInterfaceNil() bool {
	return scrProc == nil
}
