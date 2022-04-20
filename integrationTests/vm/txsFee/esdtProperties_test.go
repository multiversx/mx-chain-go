package txsFee

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestESDTPropertiesFromSCShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)

	// create an address with ESDT token
	sndAddr := []byte("12345678901234567890123456789012")

	esdtBalance := big.NewInt(100000000)
	token := []byte("TOKEN-010101")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 10, esdtBalance)
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 1, esdtBalance)
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	// deploy contract
	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	scAddress := utils.DoDeploySecond(t, testContext, "testdata/multi-transfer-esdt.wasm", ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	gasLimit := uint64(500000)
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  sndAddr,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Data:     []byte("amIFrozen@" + hex.EncodeToString(token)),
	}

	_ = logger.SetLogLevel("*:TRACE,arwen:TRACE,arwen/storage:TRACE,gasTrace:TRACE")
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)
}
