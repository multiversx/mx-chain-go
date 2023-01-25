package multiShard

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestESDTTransferShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContext.ShardCoordinator.ComputeId(sndAddr))

	rcvAddr := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContext.ShardCoordinator.ComputeId(rcvAddr))

	egldBalance := big.NewInt(100000000)
	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	gasPrice := uint64(10)
	gasLimit := uint64(40)
	tx := utils.CreateESDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100), gasPrice, gasLimit)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedReceiverBalance := big.NewInt(100)
	utils.CheckESDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)
}
