package smartContract

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-vm-common"
)

func (sc *scProcessor) CreateVMCallInput(tx *transaction.Transaction) (*vmcommon.ContractCallInput, error) {
	return sc.createVMCallInput(tx)
}

func (sc *scProcessor) CreateVMDeployInput(tx *transaction.Transaction) (*vmcommon.ContractCreateInput, error) {
	return sc.createVMDeployInput(tx)
}

func (sc *scProcessor) CreateVMInput(tx *transaction.Transaction) (*vmcommon.VMInput, error) {
	return sc.createVMInput(tx)
}

func (sc *scProcessor) ProcessVMOutput(vmOutput *vmcommon.VMOutput, tx *transaction.Transaction, acntSnd, acntDst state.AccountHandler) error {
	return sc.processVMOutput(vmOutput, tx, acntSnd, acntDst)
}

func (sc *scProcessor) RefundGasToSender(gasRefund *big.Int, tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	return sc.refundGasToSender(gasRefund, tx, acntSnd)
}

func (sc *scProcessor) ProcessSCOutputAccounts(outputAccounts []*vmcommon.OutputAccount, acntDst state.AccountHandler) error {
	return sc.processSCOutputAccounts(outputAccounts, acntDst)
}

func (sc *scProcessor) DeleteAccounts(deletedAccounts [][]byte) error {
	return sc.deleteAccounts(deletedAccounts)
}

func (sc *scProcessor) GetAccountFromAddress(address []byte) (state.AccountHandler, error) {
	return sc.getAccountFromAddress(address)
}

func (sc *scProcessor) SaveSCOutputToCurrentState(output *vmcommon.VMOutput) error {
	return sc.saveSCOutputToCurrentState(output)
}

func (sc *scProcessor) SaveReturnData(returnData []*big.Int) error {
	return sc.saveReturnData(returnData)
}

func (sc *scProcessor) SaveReturnCode(returnCode vmcommon.ReturnCode) error {
	return sc.saveReturnCode(returnCode)
}

func (sc *scProcessor) SaveLogsIntoState(logs []*vmcommon.LogEntry) error {
	return sc.saveLogsIntoState(logs)
}
