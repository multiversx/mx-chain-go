package mandosConverter

import (
	"bytes"
	"errors"
	"math/big"

	mge "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/elrondgo-exporter"
	mgutil "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/util"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var errDeployRetCodeNotOk = errors.New("returnCode is not 0(Ok)")

// CreateAccountsFromMandosAccs uses mandosAccounts to populate the AccountsAdapter
func CreateAccountsFromMandosAccs(tc *vm.VMTestContext, mandosUserAccounts []*mge.TestAccount) (err error) {
	for _, mandosAcc := range mandosUserAccounts {
		acc, err := tc.Accounts.LoadAccount(mandosAcc.GetAddress())
		if err != nil {
			return err
		}
		account := acc.(state.UserAccountHandler)
		account.IncreaseNonce(mandosAcc.GetNonce())
		err = account.AddToBalance(mandosAcc.GetBalance())
		if err != nil {
			return err
		}

		mandosAccStorage := mandosAcc.GetStorage()
		for key, value := range mandosAccStorage {
			err = account.DataTrieTracker().SaveKeyValue([]byte(key), value)
			if err != nil {
				return err
			}
		}

		accountCode := mandosAcc.GetCode()
		if len(accountCode) != 0 {
			account.SetCode(accountCode)
			ownerAddress := mandosAcc.GetOwner()
			account.SetOwnerAddress(ownerAddress)
			account.SetCodeMetadata([]byte{0, 0})
		}
		err = tc.Accounts.SaveAccount(account)
		if err != nil {
			return err
		}
	}
	_, err = tc.Accounts.Commit()
	if err != nil {
		return err
	}

	return nil
}

// CreateTransactionsFromMandosTxs converts mandos transactions intro trasnsactions that can be processed by the txProcessor
func CreateTransactionsFromMandosTxs(mandosTxs []*mge.Transaction) (transactions []*dataTransaction.Transaction) {
	transactions = make([]*dataTransaction.Transaction, 0)
	for _, mandosTx := range mandosTxs {
		gasLimit, gasPrice := mandosTx.GetGasLimitAndPrice()
		var data []byte
		esdtTransfers := mandosTx.GetESDTTransfers()
		endpointName := mandosTx.GetCallFunction()
		args := mandosTx.GetCallArguments()
		if len(esdtTransfers) != 0 {
			data = mgutil.CreateMultiTransferData(mandosTx.GetReceiverAddress(), esdtTransfers, endpointName, args)
		} else {
			data = createData(endpointName, args)
		}

		tx := vm.CreateTransaction(
			mandosTx.GetNonce(),
			mandosTx.GetCallValue(),
			mandosTx.GetSenderAddress(),
			mandosTx.GetReceiverAddress(),
			gasPrice,
			gasLimit,
			data)
		transactions = append(transactions, tx)
	}
	return transactions
}

// DeploySCsFromMandosDeployTxs deploys all smartContracts correspondent to "scDeploy" in a mandos test, then replaces with the correct computed address in all the transactions.
func DeploySCsFromMandosDeployTxs(testContext *vm.VMTestContext, deployMandosTxs []*mge.Transaction, mandosTxs []*mge.Transaction, deployedScAccounts []*mge.TestAccount) (err error) {
	for _, deployMandosTransaction := range deployMandosTxs {
		deployedScAddress, err := deploySC(testContext, deployMandosTransaction)
		if err != nil {
			return err
		}
		addressToBeReplaced := deployedScAccounts[0].GetAddress()
		for _, mandosTx := range mandosTxs {
			if bytes.Equal(mandosTx.GetReceiverAddress(), addressToBeReplaced) {
				mandosTx.WithReceiverAddress(deployedScAddress)
			}
		}
		deployedScAccounts = deployedScAccounts[1:]
	}
	return nil
}

func createData(functionName string, arguments [][]byte) []byte {
	builder := txDataBuilder.NewBuilder()
	builder.Func(functionName)
	for _, arg := range arguments {
		builder.Bytes(arg)
	}
	return builder.ToBytes()
}

func deploySC(testContext *vm.VMTestContext, deployMandosTx *mge.Transaction) (scAddress []byte, err error) {
	gasLimit, gasPrice := deployMandosTx.GetGasLimitAndPrice()
	ownerAddr := deployMandosTx.GetSenderAddress()
	deployData := deployMandosTx.GetDeployData()

	ownerAcc, err := testContext.Accounts.LoadAccount(ownerAddr)
	if err != nil {
		return nil, err
	}
	ownerNonce := ownerAcc.GetNonce()
	tx := vm.CreateTransaction(ownerNonce, big.NewInt(0), ownerAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, deployData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return nil, err
	}
	if retCode != vmcommon.Ok {
		return nil, errDeployRetCodeNotOk
	}
	_, err = testContext.Accounts.Commit()
	if err != nil {
		return nil, err
	}
	scAddress, err = testContext.BlockchainHook.NewAddress(ownerAddr, ownerNonce, factory.ArwenVirtualMachine)
	if err != nil {
		return nil, err
	}

	return scAddress, nil
}
