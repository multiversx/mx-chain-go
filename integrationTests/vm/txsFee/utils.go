package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/require"
)

var protoMarshalizer = &marshal.GogoProtoMarshalizer{}

func doDeploy(t *testing.T, testContext *vm.VMTestContext) (scAddr []byte, owner []byte) {
	owner = []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(2000)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)

	scCode := arwen.GetSCCode("../arwen/testdata/counter/output/counter.wasm")
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(arwen.CreateDeployTxData(scCode)))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(89030)
	vm.TestAccount(t, testContext.Accounts, owner, senderNonce+1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10970), accumulatedFee)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.ArwenVirtualMachine)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)

	return scAddr, owner
}

func prepareRelayerTxData(innerTx *transaction.Transaction) []byte {
	userTxBytes, _ := protoMarshalizer.Marshal(innerTx)
	return []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxBytes))
}

func checkOwnerAddr(t *testing.T, testContext *vm.VMTestContext, scAddr []byte, owner []byte) {
	acc, err := testContext.Accounts.GetExistingAccount(scAddr)
	require.Nil(t, err)

	userAcc, ok := acc.(state.UserAccountHandler)
	require.True(t, ok)

	currentOwner := userAcc.GetOwnerAddress()
	require.Equal(t, owner, currentOwner)
}
