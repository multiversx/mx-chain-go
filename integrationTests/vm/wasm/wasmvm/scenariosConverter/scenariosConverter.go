package scenariosConverter

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-scenario-go/scenario/exporter"
	scenmodel "github.com/multiversx/mx-chain-scenario-go/scenario/model"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var errReturnCodeNotOk = errors.New("returnCode is not 0(Ok)")

// CreateAccountsFromScenariosAccs uses scenariosAccounts to populate the AccountsAdapter
func CreateAccountsFromScenariosAccs(tc *vm.VMTestContext, scenariosUserAccounts []*exporter.TestAccount) error {
	for _, scenariosAcc := range scenariosUserAccounts {
		acc, err := tc.Accounts.LoadAccount(scenariosAcc.GetAddress())
		if err != nil {
			return err
		}
		account := acc.(state.UserAccountHandler)
		account.IncreaseNonce(scenariosAcc.GetNonce())
		err = account.AddToBalance(scenariosAcc.GetBalance())
		if err != nil {
			return err
		}

		scenariosAccStorage := scenariosAcc.GetStorage()
		for key, value := range scenariosAccStorage {
			err = account.SaveKeyValue([]byte(key), value)
			if err != nil {
				return err
			}
		}

		accountCode := scenariosAcc.GetCode()
		if len(accountCode) != 0 {
			account.SetCode(accountCode)
			ownerAddress := scenariosAcc.GetOwner()
			account.SetOwnerAddress(ownerAddress)
			account.SetCodeMetadata([]byte{0, 0})
		}
		err = tc.Accounts.SaveAccount(account)
		if err != nil {
			return err
		}
	}
	_, err := tc.Accounts.Commit()
	if err != nil {
		return err
	}

	return nil
}

// CreateTransactionsFromScenariosTxs converts scenarios transactions intro trasnsactions that can be processed by the txProcessor
func CreateTransactionsFromScenariosTxs(scenariosTxs []*exporter.Transaction) (transactions []*transaction.Transaction) {
	var data []byte
	transactions = make([]*transaction.Transaction, 0)

	for _, scenariosTx := range scenariosTxs {
		gasLimit, gasPrice := scenariosTx.GetGasLimitAndPrice()
		esdtTransfers := scenariosTx.GetESDTTransfers()
		endpointName := scenariosTx.GetCallFunction()
		args := scenariosTx.GetCallArguments()
		if len(esdtTransfers) != 0 {
			data = scenmodel.CreateMultiTransferData(scenariosTx.GetReceiverAddress(), esdtTransfers, endpointName, args)
		} else {
			data = createData(endpointName, args)
		}

		tx := vm.CreateTransaction(
			scenariosTx.GetNonce(),
			scenariosTx.GetCallValue(),
			scenariosTx.GetSenderAddress(),
			scenariosTx.GetReceiverAddress(),
			gasPrice,
			gasLimit,
			data)
		if len(esdtTransfers) != 0 {
			tx.RcvAddr = tx.SndAddr
		}
		transactions = append(transactions, tx)
	}
	return transactions
}

// DeploySCsFromScenariosDeployTxs deploys all smartContracts correspondent to "scDeploy" in a scenarios test, then replaces with the correct computed address in all the transactions.
func DeploySCsFromScenariosDeployTxs(testContext *vm.VMTestContext, deployScenariosTxs []*exporter.Transaction) ([][]byte, error) {
	newScAddresses := make([][]byte, 0)
	for _, deployScenariosTransaction := range deployScenariosTxs {
		deployedScAddress, err := deploySC(testContext, deployScenariosTransaction)
		if err != nil {
			return newScAddresses, err
		}
		newScAddresses = append(newScAddresses, deployedScAddress)
	}
	return newScAddresses, nil
}

// ReplaceScenariosScAddressesWithNewScAddresses corrects the Scenarios SC Addresses, with the new Addresses obtained from deploying the SCs
func ReplaceScenariosScAddressesWithNewScAddresses(deployedScAccounts []*exporter.TestAccount, newScAddresses [][]byte, scenariosTxs []*exporter.Transaction) {
	for _, newScAddr := range newScAddresses {
		addressToBeReplaced := deployedScAccounts[0].GetAddress()
		for _, scenariosTx := range scenariosTxs {
			if bytes.Equal(scenariosTx.GetReceiverAddress(), addressToBeReplaced) {
				scenariosTx.WithReceiverAddress(newScAddr)
			}
		}
		deployedScAccounts = deployedScAccounts[1:]
	}
}

func createData(functionName string, arguments [][]byte) []byte {
	builder := txDataBuilder.NewBuilder()
	builder.Func(functionName)
	for _, arg := range arguments {
		builder.Bytes(arg)
	}
	return builder.ToBytes()
}

func deploySC(testContext *vm.VMTestContext, deployScenariosTx *exporter.Transaction) (scAddress []byte, err error) {
	gasLimit, gasPrice := deployScenariosTx.GetGasLimitAndPrice()
	ownerAddr := deployScenariosTx.GetSenderAddress()
	deployData := deployScenariosTx.GetDeployData()

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
		return nil, errReturnCodeNotOk
	}
	_, err = testContext.Accounts.Commit()
	if err != nil {
		return nil, err
	}
	scAddress, err = testContext.BlockchainHook.NewAddress(ownerAddr, ownerNonce, factory.WasmVirtualMachine)
	if err != nil {
		return nil, err
	}

	return scAddress, nil
}
