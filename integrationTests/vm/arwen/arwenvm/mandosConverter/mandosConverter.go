package mandosConverter

import (
	mge "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/elrondgo-exporter"
	mgutil "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/util"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
)

func CreateAccountsFromMandosAccs(tc vm.VMTestContext, mandosUserAccounts []*mge.TestAccount) (err error) {
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

func createData(functionName string, arguments [][]byte) []byte {
	builder := txDataBuilder.NewBuilder()
	builder.Func(functionName)
	for _, arg := range arguments {
		builder.Bytes(arg)
	}
	return builder.ToBytes()
}
