package smartContract

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
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

func (sc *scProcessor) ProcessVMOutput(vmOutput *vmcommon.VMOutput, tx *transaction.Transaction, acntSnd state.AccountHandler, round uint32) ([]data.TransactionHandler, error) {
	return sc.processVMOutput(vmOutput, tx, acntSnd, round)
}

func (sc *scProcessor) RefundGasToSender(gasRefund *big.Int, tx *transaction.Transaction, txHash []byte, acntSnd state.AccountHandler) (*smartContractResult.SmartContractResult, error) {
	return sc.refundGasToSender(gasRefund, tx, txHash, acntSnd)
}

func (sc *scProcessor) ProcessSCOutputAccounts(outputAccounts []*vmcommon.OutputAccount) ([]*vmcommon.OutputAccount, error) {
	return sc.processSCOutputAccounts(outputAccounts)
}

func (sc *scProcessor) DeleteAccounts(deletedAccounts [][]byte) error {
	return sc.deleteAccounts(deletedAccounts)
}

func (sc *scProcessor) GetAccountFromAddress(address []byte) (state.AccountHandler, error) {
	return sc.getAccountFromAddress(address)
}

func (sc *scProcessor) SaveSCOutputToCurrentState(output *vmcommon.VMOutput, round uint32, txHash []byte) error {
	return sc.saveSCOutputToCurrentState(output, round, txHash)
}

func (sc *scProcessor) SaveReturnData(returnData []*big.Int, round uint32, txHash []byte) error {
	return sc.saveReturnData(returnData, round, txHash)
}

func (sc *scProcessor) SaveReturnCode(returnCode vmcommon.ReturnCode, round uint32, txHash []byte) error {
	return sc.saveReturnCode(returnCode, round, txHash)
}

func (sc *scProcessor) SaveLogsIntoState(logs []*vmcommon.LogEntry, round uint32, txHash []byte) error {
	return sc.saveLogsIntoState(logs, round, txHash)
}

func (sc *scProcessor) ProcessSCPayment(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	return sc.processSCPayment(tx, acntSnd)
}

func (sc *scProcessor) CreateCrossShardTransactions(
	crossOutAccs []*vmcommon.OutputAccount,
	tx *transaction.Transaction,
	txHash []byte,
) ([]data.TransactionHandler, error) {
	return sc.createCrossShardTransactions(crossOutAccs, tx, txHash)
}
