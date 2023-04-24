package deployment

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var senderAddressBytes = []byte("12345678901234567890123456789012")
var senderNonce = uint64(0)
var senderBalance = big.NewInt(1000000000000)

func TestScDeployShouldManageCorrectlyTheCodeMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorAndAccountsWithVMs(
		senderNonce,
		senderAddressBytes,
		senderBalance,
		config.EnableEpochs{
			IsPayableBySCEnableEpoch: 1,
			SetGuardianEnableEpoch:   1,
		},
	)
	require.Nil(t, err)
	defer testContext.Close()

	t.Run("payable by SC is not active, should set that flag to false: backwards compatibility reasons", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "FFFF")

		expectedCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: false,
			Upgradeable: true,
			Readable:    true,
			Guarded:     false,
		}

		assert.Equal(t, expectedCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("payable by SC is active, should return the correct code metadata containing the PayableBySC flag value", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "FFFF")

		expectedCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: true,
			Upgradeable: true,
			Readable:    true,
			Guarded:     false,
		}

		assert.Equal(t, expectedCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("payable by SC is active, wrong number of bytes in code metadata argument will return an empty code metadata value: backwards compatibility reasons", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "FFFFFF")

		assert.Equal(t, []byte{0, 0}, getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("SC deploys child SC while payable by SC is inactive with a bad code metadata should work: backwards compatibility reasons", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0})

		goodCodeMetadata := []byte{05, 02}
		badCodeMetadata := []byte{0xFF, 0xFF}
		scParent, scChild, retCode, errDeploy := deployChildAndParentContracts(t, testContext, senderAddressBytes, hex.EncodeToString(goodCodeMetadata), hex.EncodeToString(badCodeMetadata))

		assert.Equal(t, goodCodeMetadata, getCodeMetadata(t, testContext.Accounts, scParent))
		assert.Equal(t, badCodeMetadata, getCodeMetadata(t, testContext.Accounts, scChild))
		assert.Equal(t, vmcommon.Ok, retCode)
		assert.Nil(t, errDeploy)
	})
	t.Run("SC deploys child SC while payable by SC is active should handle the child SC deploy", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		goodCodeMetadata := []byte{05, 02}
		t.Run("bad code metadata by value", func(t *testing.T) {
			badCodeMetadata := []byte{0xFF, 0xFF}
			scParent, scChild, retCode, errDeploy := deployChildAndParentContracts(t, testContext, senderAddressBytes, hex.EncodeToString(goodCodeMetadata), hex.EncodeToString(badCodeMetadata))

			assert.Equal(t, goodCodeMetadata, getCodeMetadata(t, testContext.Accounts, scParent))
			assert.False(t, accountExists(testContext.Accounts, scChild))
			assert.Equal(t, vmcommon.ExecutionFailed, retCode)
			assert.Nil(t, errDeploy)
		})
		t.Run("bad code metadata by number of values", func(t *testing.T) {
			badCodeMetadata := []byte{0xFF, 0xFF, 0xFF}
			scParent, scChild, retCode, errDeploy := deployChildAndParentContracts(t, testContext, senderAddressBytes, hex.EncodeToString(goodCodeMetadata), hex.EncodeToString(badCodeMetadata))

			assert.Equal(t, goodCodeMetadata, getCodeMetadata(t, testContext.Accounts, scParent))
			assert.False(t, accountExists(testContext.Accounts, scChild))
			assert.Equal(t, vmcommon.ExecutionFailed, retCode)
			assert.Nil(t, errDeploy)
		})
		t.Run("ok code metadata should work", func(t *testing.T) {
			scParent, scChild, retCode, errDeploy := deployChildAndParentContracts(t, testContext, senderAddressBytes, hex.EncodeToString(goodCodeMetadata), hex.EncodeToString(goodCodeMetadata))

			assert.Equal(t, goodCodeMetadata, getCodeMetadata(t, testContext.Accounts, scParent))
			assert.Equal(t, goodCodeMetadata, getCodeMetadata(t, testContext.Accounts, scChild))
			assert.Equal(t, vmcommon.Ok, retCode)
			assert.Nil(t, errDeploy)
		})
	})
}

func deploySC(tb testing.TB, testContext *vm.VMTestContext, senderAddressBytes []byte, codeMetadataHex string, codeHex string) []byte {
	transferOnCalls := big.NewInt(50)

	accnt, err := testContext.Accounts.LoadAccount(senderAddressBytes)
	require.Nil(tb, err)

	nonce := accnt.GetNonce()
	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		nonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxDataWithCodeMetadata(codeHex, codeMetadataHex),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(tb, err)
	require.Equal(tb, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	vmTypeBytes, _ := hex.DecodeString(wasm.VMTypeHex)
	contractAddressBytes, err := testContext.BlockchainHook.NewAddress(senderAddressBytes, nonce, vmTypeBytes)
	require.Nil(tb, err)

	return contractAddressBytes
}

func deployDummySCReturningContractAddress(tb testing.TB, testContext *vm.VMTestContext, senderAddressBytes []byte, codeMetadataHex string) []byte {
	scCode := wasm.GetSCCode("../../testdata/misc/fib_wasm/output/fib_wasm.wasm")

	return deploySC(tb, testContext, senderAddressBytes, codeMetadataHex, scCode)
}

func deployChildAndParentContracts(
	tb testing.TB,
	testContext *vm.VMTestContext,
	senderAddressBytes []byte,
	codeMetadataParentHex string,
	codeMetadataChildHex string,
) ([]byte, []byte, vmcommon.ReturnCode, error) {
	transferOnCalls := big.NewInt(50)

	parentScCode := wasm.GetSCCode("../../testdata/deployer-custom/output/deployer-custom.wasm")
	scParentAddressBytes := deploySC(tb, testContext, senderAddressBytes, codeMetadataParentHex, parentScCode)

	scParentAccnt, err := testContext.Accounts.LoadAccount(scParentAddressBytes)
	require.Nil(tb, err)
	parentNonce := scParentAccnt.GetNonce()

	childScCode := wasm.GetSCCode("../../testdata/counter/output/counter.wasm")
	accnt, err := testContext.Accounts.LoadAccount(senderAddressBytes)
	require.Nil(tb, err)

	nonce := accnt.GetNonce()
	tx := vm.CreateTx(
		senderAddressBytes,
		scParentAddressBytes,
		nonce,
		transferOnCalls,
		gasPrice,
		gasLimit*10,
		fmt.Sprintf("%s@%s@%s", "deployChild", childScCode, codeMetadataChildHex),
	)

	vmTypeBytes, _ := hex.DecodeString(wasm.VMTypeHex)
	scChildAddressBytes, err := testContext.BlockchainHook.NewAddress(scParentAddressBytes, parentNonce, vmTypeBytes)
	require.Nil(tb, err)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil || returnCode != vmcommon.Ok {
		return scParentAddressBytes, scChildAddressBytes, returnCode, err
	}
	require.Nil(tb, err)
	require.Equal(tb, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	return scParentAddressBytes, scChildAddressBytes, vmcommon.Ok, nil
}

func getCodeMetadata(tb testing.TB, accounts state.AccountsAdapter, address []byte) []byte {
	account, err := accounts.LoadAccount(address)
	require.Nil(tb, err)

	userAccount := account.(state.UserAccountHandler)
	return userAccount.GetCodeMetadata()
}

func accountExists(accounts state.AccountsAdapter, address []byte) bool {
	_, err := accounts.GetExistingAccount(address)

	return err == nil
}
