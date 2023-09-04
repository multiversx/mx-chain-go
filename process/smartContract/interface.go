package smartContract

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type SCProcessorBaseHandler interface {
	process.SmartContractProcessor
	process.SmartContractResultProcessor

	ArgsParser() process.ArgumentsParser
	TxTypeHandler() process.TxTypeHandler

	CheckSCRBeforeProcessing(scr *smartContractResult.SmartContractResult) (process.ScrProcessingDataHandler, error)
	ExecuteBuiltInFunction(tx data.TransactionHandler, acntSnd, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error)
	ProcessIfError(acntSnd state.UserAccountHandler,
		txHash []byte,
		tx data.TransactionHandler,
		returnCode string,
		returnMessage []byte,
		snapshot int,
		gasLocked uint64,
	) error

	IsInterfaceNil() bool
}
