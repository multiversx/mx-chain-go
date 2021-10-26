package Benchmark_TestEllipticCurveScalarMultP224

import (
	"fmt"
	"testing"

	mge "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mandos-go/elrondgo-exporter"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	mc "github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm/mandosConverter"
	"github.com/stretchr/testify/require"
)

func TestMandosConverter_EllipticCurves(t *testing.T) {
	mandosAccounts, mandosTransactions, err := mge.GetAccountsAndTransactionsFromMandos("./elliptic_curves.scen.json")
	require.Nil(t, err)
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	err = mc.CreateAccountsFromMandosAccs(*testContext, mandosAccounts)
	require.Nil(t, err)
	mc.CheckAccounts(t, testContext.Accounts, mandosAccounts)
	transactions := mc.CreateTransactionsFromMandosTxs(mandosTransactions)
	mc.CheckTransactions(t, transactions, mandosTransactions)
}

func Benchmark_MandosConverter_EllipticCurves(b *testing.B) {
	testContext, transactions, err := mc.SetStateFromMandosTest("./elliptic_curves.scen.json")
	if err != nil {
		fmt.Println("Setting state went wrong: ", err)
		return
	}
	b.ResetTimer()
	err = mc.RunSingleTransactionBenchmark(b, testContext, transactions[0])
	if err != nil {
		fmt.Println("Proccess transaction went wrong: ", err)
	}
}
