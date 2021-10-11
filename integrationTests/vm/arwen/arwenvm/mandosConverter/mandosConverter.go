package mandosConverter

import (
	ame "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/arwenmandos/elrondgo-exporter"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
)

func createAccountsFromMandosAccs(accAdapter state.AccountsAdapter,mandosAccounts []*ame.TestAccount) (err error) {
	for _, mandosAcc := range mandosAccounts {
		account, err := accAdapter.LoadAccount(mandosAcc.GetAddress())
		if err != nil  {
			return err
		}
		userAccount := account.(state.UserAccountHandler)
		userAccount.IncreaseNonce(mandosAcc.GetNonce())
		err = userAccount.AddToBalance(mandosAcc.GetBalance())
		if err != nil {
			return err
		}

		mandosAccStorage := mandosAcc.GetStorage()
		for key,value := range mandosAccStorage {
			err = userAccount.DataTrieTracker().SaveKeyValue([]byte(key),value)
			if err != nil {
				return err
			}
		}
		err = accAdapter.SaveAccount(account)
		if err != nil {
			return err
		}
	}
	_,err = accAdapter.Commit()
	if err!=nil {
		return err
	}
	return nil
}

func breakMandosTxsIntoDeploysAndTransactions(mandosTxs []*ame.Transaction) (deployTxs []*dataTransaction.Transaction, transactions []*dataTransaction.Transaction) {
	transactions = make([]*dataTransaction.Transaction,0)
	deployTxs = make([]*dataTransaction.Transaction,0)
	for _,mandosTx := range mandosTxs {
		gasLimit,gasPrice := mandosTx.GetGasLimitAndPrice()
		if mandosTx.IsDeploy() {
			scCode := arwen.GetSCCode(mandosTx.GetDeployPath())
			tx := vm.CreateTransaction(
				mandosTx.GetNonce(),
				mandosTx.GetCallValue(),
				mandosTx.GetSenderAddress(),
				vm.CreateEmptyAddress(),
				gasPrice,
				gasLimit,
				[]byte(arwen.CreateDeployTxData(scCode)))
			deployTxs = append(deployTxs,   tx)
		} else {
			data := createData(mandosTx.GetCallFunction(),mandosTx.GetCallArguments())
			tx := vm.CreateTransaction(
				mandosTx.GetNonce(),
				mandosTx.GetCallValue(),
				mandosTx.GetSenderAddress(),
				mandosTx.GetReceiverAddress(),
				gasPrice,
				gasLimit,
				data)
			transactions = append(transactions,tx)
		}
	}
	return deployTxs,transactions
}


func createData(functionName string, arguments [][]byte) []byte {
	builder := txDataBuilder.NewBuilder()
	builder.Func(functionName)
	for _,arg := range arguments {
		builder.Bytes(arg)
	}
	return builder.ToBytes()
}
