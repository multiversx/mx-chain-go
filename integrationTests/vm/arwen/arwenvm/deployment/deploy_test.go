package deployment

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScDeployShouldManageCorrectlyTheCodeMetadata(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{
			IsPayableBySCEnableEpoch: 1,
		},
	)
	require.Nil(t, err)
	defer testContext.Close()

	t.Run("payable by SC is not active", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "FFFF")

		expectedCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: false,
			Upgradeable: true,
			Readable:    true,
		}

		assert.Equal(t, expectedCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("payable by SC is active", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "FFFF")

		expectedCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: true,
			Upgradeable: true,
			Readable:    true,
		}

		assert.Equal(t, expectedCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("payable by SC is active, wrong number of bytes in code metadata argument", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "FFFFFF")

		assert.Equal(t, []byte{0, 0}, getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
}

func deployDummySCReturningContractAddress(tb testing.TB, testContext *vm.VMTestContext, senderAddressBytes []byte, codeMetadataHex string) []byte {
	gasPrice := uint64(1)
	gasLimit := uint64(1000)
	transferOnCalls := big.NewInt(50)

	accnt, err := testContext.Accounts.LoadAccount(senderAddressBytes)
	require.Nil(tb, err)

	nonce := accnt.GetNonce()
	scCode := arwen.GetSCCode("../../testdata/misc/fib_arwen/output/fib_arwen.wasm")
	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		nonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxDataWithCodeMetadata(scCode, codeMetadataHex),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(tb, err)
	require.Equal(tb, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	vmTypeBytes, _ := hex.DecodeString(arwen.VMTypeHex)
	contractAddressBytes, err := testContext.BlockchainHook.NewAddress(senderAddressBytes, nonce, vmTypeBytes)
	require.Nil(tb, err)

	return contractAddressBytes
}

func getCodeMetadata(tb testing.TB, accounts state.AccountsAdapter, address []byte) []byte {
	account, err := accounts.LoadAccount(address)
	require.Nil(tb, err)

	userAccount := account.(state.UserAccountHandler)
	return userAccount.GetCodeMetadata()
}
