package processorV2

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/process"
)

// SovereignSCProcessArgs - arguments for creating a new sovereign smart contract processor
type SovereignSCProcessArgs struct {
	ArgsParser               process.ArgumentsParser
	TxTypeHandler            process.TxTypeHandler
	SCProcessorHelperHandler process.SCProcessorHelperHandler
	SmartContractProcessor   process.SmartContractProcessorFacade
}

type sovereignSCProcessor struct {
	process.SmartContractProcessorFacade

	argsParser        process.ArgumentsParser
	txTypeHandler     process.TxTypeHandler
	scProcessorHelper process.SCProcessorHelperHandler
}

// NewSovereignSCRProcessor creates a sovereign scr processor
func NewSovereignSCRProcessor(args SovereignSCProcessArgs) (*sovereignSCProcessor, error) {
	if check.IfNil(args.SmartContractProcessor) {
		return nil, process.ErrNilSmartContractResultProcessor
	}
	if check.IfNil(args.ArgsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.SCProcessorHelperHandler) {
		return nil, process.ErrNilSCProcessorHelper
	}

	if check.IfNil(args.ArgsParser) {
		return nil, process.ErrNilArgumentParser
	}

	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}

	return &sovereignSCProcessor{
		SmartContractProcessorFacade: args.SmartContractProcessor,
		argsParser:                   args.ArgsParser,
		txTypeHandler:                args.TxTypeHandler,
		scProcessorHelper:            args.SCProcessorHelperHandler,
	}, nil
}

// ProcessSmartContractResult updates the account state from the smart contract result
func (sc *sovereignSCProcessor) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	if check.IfNil(scr) {
		return 0, process.ErrNilSmartContractResult
	}

	log.Trace("sovereignSCProcessor.ProcessSmartContractResult()", "sender", scr.GetSndAddr(), "receiver", scr.GetRcvAddr(), "data", string(scr.GetData()))

	var err error
	returnCode := vmcommon.UserError
	if !bytes.Equal(scr.SndAddr, core.ESDTSCAddress) {
		return returnCode, fmt.Errorf("%w, expected ESDTSCAddress", errInvalidSenderAddress)
	}

	scrData, err := sc.scProcessorHelper.CheckSCRBeforeProcessing(scr)
	if err != nil {
		return returnCode, err
	}

	txType, _, _ := sc.txTypeHandler.ComputeTransactionType(scr)
	switch txType {
	case process.BuiltInFunctionCall:
		err = sc.checkBuiltInFuncCall(string(scr.Data))
		if err != nil {
			return returnCode, err
		}

		return sc.ExecuteBuiltInFunction(scr, nil, scrData.GetDestination())
	default:
		err = process.ErrWrongTransaction
	}

	return returnCode, sc.ProcessIfError(scrData.GetSender(), scrData.GetHash(), scr, err.Error(), scr.ReturnMessage, scrData.GetSnapshot(), 0)
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
