package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
)

// SmartContractResultsProcessorMock -
type SmartContractResultsProcessorMock struct {
	ProcessSmartContractResultCalled func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error)
}

// ProcessSmartContractResult -
func (scrp *SmartContractResultsProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	if scrp.ProcessSmartContractResultCalled == nil {
		return 0, nil
	}

	return scrp.ProcessSmartContractResultCalled(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scrp *SmartContractResultsProcessorMock) IsInterfaceNil() bool {
	return scrp == nil
}
