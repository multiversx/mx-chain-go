package testadder

import (
	"fmt"
	"testing"

	mge "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/elrondgo-exporter"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	mc "github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm/mandosConverter"
	"github.com/stretchr/testify/require"
)

func TestMandosConverter_AdderWithExternalSteps(t *testing.T) {
	mandosAccounts, _, mandosTransactions, _, err := mge.GetAccountsAndTransactionsFromMandos("./adder_with_external_steps.scen.json")
	require.Nil(t, err)
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	err = mc.CreateAccountsFromMandosAccs(testContext, mandosAccounts)
	require.Nil(t, err)
	mc.CheckAccounts(t, testContext.Accounts, mandosAccounts)
	transactions := mc.CreateTransactionsFromMandosTxs(mandosTransactions)
	mc.CheckTransactions(t, transactions, mandosTransactions)
}

func Benchmark_MandosConverter_SimpleAdderWithDeploy(b *testing.B) {
	testContext, transactions, err := mc.SetStateFromMandosTest("./adder.scen.json")
	if err != nil {
		fmt.Println("Setting state went wrong: ", err)
		return
	}
	defer testContext.Close()
	mc.RunSingleTransactionBenchmark(b, testContext, transactions[0])
}

func Benchmark_MandosConverter_AdderWithExternalSteps(b *testing.B) {
	testContext, transactions, err := mc.SetStateFromMandosTest("./adder_with_external_steps.scen.json")
	if err != nil {
		fmt.Println("Setting state went wrong: ", err)
		return
	}
	defer testContext.Close()
	mc.RunSingleTransactionBenchmark(b, testContext, transactions[0])
}
