package mandosConverter

import (
	"testing"

	mge "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/elrondgo-exporter"
	mgutil "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/util"
	"github.com/ElrondNetwork/elrond-go/config"

	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("mandosConverte")

// CheckAccounts will verify if mandosAccounts correspond to AccountsAdapter accounts
func CheckAccounts(t *testing.T, accAdapter state.AccountsAdapter, mandosAccounts []*mge.TestAccount) {
	for _, mandosAcc := range mandosAccounts {
		accHandler, err := accAdapter.LoadAccount(mandosAcc.GetAddress())
		require.Nil(t, err)
		account := accHandler.(state.UserAccountHandler)

		require.Equal(t, mandosAcc.GetBalance(), account.GetBalance())
		require.Equal(t, mandosAcc.GetNonce(), account.GetNonce())

		scOwnerAddress := mandosAcc.GetOwner()
		if len(scOwnerAddress) == 0 {
			require.Nil(t, account.GetOwnerAddress())
		} else {
			require.Equal(t, mandosAcc.GetOwner(), account.GetOwnerAddress())
		}

		codeHash := account.GetCodeHash()
		code := accAdapter.GetCode(codeHash)
		require.Equal(t, len(mandosAcc.GetCode()), len(code))

		mandosAccStorage := mandosAcc.GetStorage()
		CheckStorage(t, account, mandosAccStorage)
	}
}

// CheckStorage checks if the dataTrie of an account equals with the storage of the corresponding mandosAccount
func CheckStorage(t *testing.T, dataTrie state.UserAccountHandler, mandosAccStorage map[string][]byte) {
	for key := range mandosAccStorage {
		dataTrieValue, err := dataTrie.RetrieveValue([]byte(key))
		require.Nil(t, err)
		if len(mandosAccStorage[key]) == 0 {
			require.Nil(t, dataTrieValue)
		} else {
			require.Equal(t, mandosAccStorage[key], dataTrieValue)
		}
	}
}

// CheckTransactions checks if the transactions correspond with the mandosTransactions
func CheckTransactions(t *testing.T, transactions []*transaction.Transaction, mandosTransactions []*mge.Transaction) {
	expectedLength := len(mandosTransactions)
	require.Equal(t, expectedLength, len(transactions))
	for i := 0; i < expectedLength; i++ {
		expectedSender := mandosTransactions[i].GetSenderAddress()
		expectedReceiver := mandosTransactions[i].GetReceiverAddress()
		expectedCallValue := mandosTransactions[i].GetCallValue()
		expectedCallFunction := mandosTransactions[i].GetCallFunction()
		expectedCallArguments := mandosTransactions[i].GetCallArguments()
		expectedEsdtTransfers := mandosTransactions[i].GetESDTTransfers()
		expectedGasLimit, expectedGasPrice := mandosTransactions[i].GetGasLimitAndPrice()
		expectedNonce := mandosTransactions[i].GetNonce()

		require.Equal(t, expectedSender, transactions[i].GetSndAddr())
		require.Equal(t, expectedCallValue, transactions[i].GetValue())
		require.Equal(t, expectedGasLimit, transactions[i].GetGasLimit())
		require.Equal(t, expectedGasPrice, transactions[i].GetGasPrice())
		require.Equal(t, expectedNonce, transactions[i].GetNonce())

		var expectedData []byte
		if len(expectedEsdtTransfers) != 0 {
			expectedData = mgutil.CreateMultiTransferData(expectedReceiver, expectedEsdtTransfers, expectedCallFunction, expectedCallArguments)
			require.Equal(t, expectedSender, transactions[i].GetRcvAddr())
		} else {
			require.Equal(t, expectedReceiver, transactions[i].GetRcvAddr())
			expectedData = createData(expectedCallFunction, expectedCallArguments)
		}

		actualData := transactions[i].GetData()
		require.Equal(t, expectedData, actualData)
	}
}

// BenchmarkMandosSpecificTx -
func BenchmarkMandosSpecificTx(b *testing.B, mandosTestPath string) {
	testContext, transactions, benchmarkTxPos, err := SetStateFromMandosTest(mandosTestPath)
	if err != nil {
		log.Trace("Setting state went wrong:", "error", err)
		return
	}
	defer testContext.Close()
	if benchmarkTxPos == mge.InvalidBenchmarkTxPos {
		log.Trace("no transactions marked for benchmarking")
	}
	if len(transactions) > 1 {
		err = ProcessAllTransactions(testContext, transactions[:benchmarkTxPos])
		if err != nil {
			log.Trace("Processing transactions went wrong:", "error", err)
		}
	}

	err = RunSingleTransactionBenchmark(b, testContext, transactions[benchmarkTxPos])
	if err != nil {
		log.Trace("Processing benchmark transaction went wrong:", "error", err)
	}
}

// SetStateFromMandosTest recieves path to mandosTest, returns a VMTestContext with the specified accounts, an array with the specified transactions and an error
func SetStateFromMandosTest(mandosTestPath string) (testContext *vm.VMTestContext, transactions []*transaction.Transaction, bechmarkTxPos int, err error) {
	stateAndBenchmarkInfo, err := mge.GetAccountsAndTransactionsFromMandos(mandosTestPath)
	if err != nil {
		return nil, nil, mge.InvalidBenchmarkTxPos, err
	}
	testContext, err = vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	if err != nil {
		return nil, nil, mge.InvalidBenchmarkTxPos, err
	}
	err = CreateAccountsFromMandosAccs(testContext, stateAndBenchmarkInfo.Accs)
	if err != nil {
		return nil, nil, mge.InvalidBenchmarkTxPos, err
	}
	newAddresses, err := DeploySCsFromMandosDeployTxs(testContext, stateAndBenchmarkInfo.DeployTxs)
	if err != nil {
		return nil, nil, mge.InvalidBenchmarkTxPos, err
	}
	ReplaceMandosScAddressesWithNewScAddresses(stateAndBenchmarkInfo.DeployedAccs, newAddresses, stateAndBenchmarkInfo.Txs)
	transactions = CreateTransactionsFromMandosTxs(stateAndBenchmarkInfo.Txs)
	return testContext, transactions, stateAndBenchmarkInfo.BenchmarkTxPos, nil
}

// CheckConverter -
func CheckConverter(t *testing.T, mandosTestPath string) {
	stateAndBenchmarkInfo, err := mge.GetAccountsAndTransactionsFromMandos(mandosTestPath)
	require.Nil(t, err)
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	err = CreateAccountsFromMandosAccs(testContext, stateAndBenchmarkInfo.Accs)
	require.Nil(t, err)
	CheckAccounts(t, testContext.Accounts, stateAndBenchmarkInfo.Accs)
	transactions := CreateTransactionsFromMandosTxs(stateAndBenchmarkInfo.Txs)
	CheckTransactions(t, transactions, stateAndBenchmarkInfo.Txs)
}

// ProcessAllTransactions -
func ProcessAllTransactions(testContext *vm.VMTestContext, transactions []*transaction.Transaction) error {
	for _, tx := range transactions {
		sndrAccHandler, err := testContext.Accounts.LoadAccount(tx.SndAddr)
		if err != nil {
			return err
		}
		sndrAcc := sndrAccHandler.(state.UserAccountHandler)
		tx.Nonce = sndrAcc.GetNonce()
		returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
		if err != nil {
			return err
		} else if returnCode != vmcommon.Ok {
			return errReturnCodeNotOk
		}
	}
	return nil
}

// RunSingleTransactionBenchmark receives the VMTestContext (which can be created with SetStateFromMandosTest), a tx and performs a benchmark on that specific tx. If processing transaction fails, it will return error, else will return nil
func RunSingleTransactionBenchmark(b *testing.B, testContext *vm.VMTestContext, tx *transaction.Transaction) error {
	var returnCode vmcommon.ReturnCode
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
		tx.Nonce++
	}
	b.StopTimer()
	if err != nil {
		return err
	}
	if returnCode != vmcommon.Ok {
		return errReturnCodeNotOk
	}
	return nil
}
