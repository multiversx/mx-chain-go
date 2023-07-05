package processorV2

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type sovereignSCProcessor struct {
	*scProcessor
}

// TODO: use scrProcessorV2 when feat/vm1.5 is merged into feat/chain-sdk-go

// NewSovereignSCRProcessor creates a sovereign scr processor
func NewSovereignSCRProcessor(scrProc *scProcessor) (*sovereignSCProcessor, error) {
	if check.IfNil(scrProc) {
		return nil, process.ErrNilSmartContractResultProcessor
	}

	return &sovereignSCProcessor{
		scProcessor: scrProc,
	}, nil
}

// ProcessSmartContractResult updates the account state from the smart contract result
func (sc *sovereignSCProcessor) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	if check.IfNil(scr) {
		return 0, process.ErrNilSmartContractResult
	}

	log.Debug("sovereignSCProcessor.ProcessSmartContractResult()", "sender", scr.GetSndAddr(), "receiver", scr.GetRcvAddr(), "data", string(scr.GetData()))

	var err error
	returnCode := vmcommon.UserError
	if !bytes.Equal(scr.SndAddr, core.ESDTSCAddress) {
		return returnCode, fmt.Errorf("%w, expected ESDTSCAddress", errInvalidSenderAddress)
	}

	scrData, err := sc.scrChecker.CheckSCRBeforeProcessing(scr)
	if err != nil {
		return returnCode, err
	}

	txType, _ := sc.txTypeHandler.ComputeTransactionType(scr)
	switch txType {
	case process.BuiltInFunctionCall:
		err = sc.checkBuiltInFuncCall(string(scr.Data))
		if err != nil {
			return returnCode, err
		}

		return sc.ExecuteBuiltInFunction(scr, nil, scrData.Destination)
	default:
		err = process.ErrWrongTransaction
	}

	return returnCode, sc.ProcessIfError(scrData.Sender, scrData.Hash, scr, err.Error(), scr.ReturnMessage, scrData.Snapshot, 0)
}

func (sc *sovereignSCProcessor) checkBuiltInFuncCall(scrData string) error {
	function, _, err := sc.argsParser.ParseCallData(scrData)
	if err != nil {
		return err
	}

	if function != core.BuiltInFunctionMultiESDTNFTTransfer {
		return fmt.Errorf("%w, expected %s", errInvalidBuiltInFunctionCall, core.BuiltInFunctionMultiESDTNFTTransfer)
	}

	return nil
}
