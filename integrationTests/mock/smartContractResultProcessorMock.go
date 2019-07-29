package mock

import "github.com/ElrondNetwork/elrond-go/data/smartContractResult"

type SmartContractResultsProcessorMock struct {
	ProcessSmartContractResultCalled func(scr *smartContractResult.SmartContractResult) error
}

func (scrp *SmartContractResultsProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	if scrp.ProcessSmartContractResultCalled == nil {
		return nil
	}

	return scrp.ProcessSmartContractResultCalled(scr)
}
