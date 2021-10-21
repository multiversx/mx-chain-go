package mandosConverter

import (
	"testing"

	mge "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/elrondgo-exporter"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/require"
)

// CheckAccounts will verify if mandosAccounts correspond to AccountsAdapter accounts
func CheckAccounts(t *testing.T, accAdapter state.AccountsAdapter, mandosAccounts []*mge.TestAccount) {
	for _, mandosAcc := range mandosAccounts {
		account, err := accAdapter.LoadAccount(mandosAcc.GetAddress())
		require.Nil(t, err)
		ownerAccount := account.(state.UserAccountHandler)

		require.Equal(t, mandosAcc.GetBalance(), ownerAccount.GetBalance())
		require.Equal(t, mandosAcc.GetNonce(), ownerAccount.GetNonce())
		owner := mandosAcc.GetOwner()
		if len(owner) == 0 {
			require.Nil(t, ownerAccount.GetOwnerAddress())
		} else {
			require.Equal(t, mandosAcc.GetOwner(), ownerAccount.GetOwnerAddress())
		}
		mandosAccStorage := mandosAcc.GetStorage()
		accStorage := ownerAccount.DataTrieTracker()
		CheckStorage(t, accStorage, mandosAccStorage)
	}
}

// CheckStorage checks if the dataTrie of an account equals with the storage of the corresponding mandosAccount
func CheckStorage(t *testing.T, dataTrie state.DataTrieTracker, mandosAccStorage map[string][]byte) {
	for key := range mandosAccStorage {
		dataTrieValue, err := dataTrie.RetrieveValue([]byte(key))
		require.Nil(t, err)
		require.Equal(t, mandosAccStorage[key], dataTrieValue)
	}
}

// CheckTransactions checks if the transactions correspond with the mandosTransactions
func CheckTransactions(t *testing.T, transactions []*dataTransaction.Transaction, mandosTransactions []*mge.Transaction) {
	expectedLength := len(mandosTransactions)
	require.Equal(t, expectedLength, len(transactions))
	for i := 0; i < expectedLength; i++ {
		expectedSender := mandosTransactions[i].GetSenderAddress()
		expectedReceiver := mandosTransactions[i].GetReceiverAddress()
		expectedCallValue := mandosTransactions[i].GetCallValue()
		expectedCallFunction := mandosTransactions[i].GetCallFunction()
		expectedCallArguments := mandosTransactions[i].GetCallArguments()
		expectedGasLimit, expectedGasPrice := mandosTransactions[i].GetGasLimitAndPrice()
		expectedNonce := mandosTransactions[i].GetNonce()
		//expectedEsdtTransfers := mandosTransactions[i].GetESDTTransfers()

		require.Equal(t, expectedSender, transactions[i].GetSndAddr())
		require.Equal(t, expectedReceiver, transactions[i].GetRcvAddr())
		require.Equal(t, expectedCallValue, transactions[i].GetValue())
		require.Equal(t, expectedGasLimit, transactions[i].GetGasLimit())
		require.Equal(t, expectedGasPrice, transactions[i].GetGasPrice())
		require.Equal(t, expectedNonce, transactions[i].GetNonce())

		expectedData := createData(expectedCallFunction, expectedCallArguments)
		actualData := transactions[i].GetData()
		require.Equal(t, expectedData, actualData)
	}
}

func SetStateFromMandosTest(mandosTestPath string) (testContext *vm.VMTestContext,transactions []*dataTransaction.Transaction,err error) {
	mandosAccounts, mandosTransactions, err := mge.GetAccountsAndTransactionsFromMandos(mandosTestPath)
	if err != nil {
		return nil,nil, err
	}
	testContext, err = vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	if err != nil {
		return nil, nil,err
	}
	err = CreateAccountsFromMandosAccs(testContext.Accounts, mandosAccounts)
	if err != nil {
		return nil, nil,err
	}
	transactions = CreateTransactionsFromMandosTxs(mandosTransactions)
	return testContext,transactions,nil
}

func RunSingleTransactionBenchmark(b *testing.B,testContext *vm.VMTestContext, tx *dataTransaction.Transaction) (err error){
	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	b.StopTimer()
	if err!=nil {
		return err
	}
	return nil
}
