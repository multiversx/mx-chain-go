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
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gasPrice = uint64(1)
const gasLimit = uint64(10000000)

func TestScUpgradeShouldManageCorrectlyTheCodeMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

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

	t.Run("payable by SC is not active, code metadata is not parsed: backwards compatibility reasons", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "0502")

		expectedCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: false,
			Upgradeable: true,
			Readable:    true,
		}

		assert.Equal(t, expectedCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))

		newCodeMetadata := []byte{0xFF, 0xFF, 0xFF}
		returnCode, errUpgrade := upgradeDummySC(t, testContext, senderAddressBytes, contractAddress, hex.EncodeToString(newCodeMetadata))

		assert.Nil(t, errUpgrade)
		assert.Equal(t, vmcommon.Ok, returnCode)
		assert.Equal(t, newCodeMetadata, getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("payable by SC is active, wrong code metadata should error", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "0502")

		deployCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: false,
			Upgradeable: true,
			Readable:    true,
		}
		assert.Equal(t, deployCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))

		t.Run("valid number of bytes, invalid code", func(t *testing.T) {
			newCodeMetadata := "FFFF"
			returnCode, errUpgrade := upgradeDummySC(t, testContext, senderAddressBytes, contractAddress, newCodeMetadata)
			assert.Nil(t, errUpgrade)
			assert.Equal(t, vmcommon.ExecutionFailed, returnCode)
			// code metadata should have not changed
			assert.Equal(t, deployCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))
		})
		t.Run("invalid number of bytes", func(t *testing.T) {
			newCodeMetadata := "FFFFFF"
			returnCode, errUpgrade := upgradeDummySC(t, testContext, senderAddressBytes, contractAddress, newCodeMetadata)
			assert.Nil(t, errUpgrade)
			assert.Equal(t, vmcommon.ExecutionFailed, returnCode)
			// code metadata should have not changed
			assert.Equal(t, deployCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))
		})
	})
	t.Run("contract has an invalid flag due to a wrong upgradeContract call. "+
		"Call the contract after the epoch for `fixing` the code metadata is enabled. Should not alter codemetada: backwards compatibility reasons", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "0502")
		deployCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: false,
			Upgradeable: true,
			Readable:    true,
		}
		assert.Equal(t, deployCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))

		newCodeMetadata := []byte{0xFF, 0xFF}
		returnCode, errUpgrade := upgradeDummySC(t, testContext, senderAddressBytes, contractAddress, hex.EncodeToString(newCodeMetadata))
		assert.Nil(t, errUpgrade)
		assert.Equal(t, vmcommon.Ok, returnCode)

		assert.Equal(t, newCodeMetadata, getCodeMetadata(t, testContext.Accounts, contractAddress))

		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		accnt, errLoad := testContext.Accounts.LoadAccount(senderAddressBytes)
		require.Nil(t, errLoad)
		nonce := accnt.GetNonce()
		tx := vm.CreateTx(senderAddressBytes, contractAddress, nonce, big.NewInt(0), gasPrice, gasLimit, "increment")
		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		require.Equal(t, returnCode, vmcommon.Ok)

		// test the code is still the newCodeMetadata value (0xFFFF) even if the last call altered the SC account's data
		assert.Equal(t, newCodeMetadata, getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("SC deploys child while payable by SC is inactive and upgrade will work with a bad code metadata: backwards compatibility reasons", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0})

		codeMetadata := vmcommon.CodeMetadata{
			Upgradeable: true,
		}
		codeMetadataHex := hex.EncodeToString(codeMetadata.ToBytes())
		badCodeMetadata := []byte{0xFF, 0xFF}
		scParent, scChild, retCodeDeploy, errDeploy := deployChildAndParentContracts(t, testContext, senderAddressBytes, codeMetadataHex, codeMetadataHex)
		assert.Equal(t, vmcommon.Ok, retCodeDeploy)
		assert.Nil(t, errDeploy)

		retCodeUpgrade, errUpgrade := upgradeChildContract(t, testContext, senderAddressBytes, scParent, scChild, hex.EncodeToString(badCodeMetadata))

		assert.Equal(t, codeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, scParent))
		assert.Equal(t, badCodeMetadata, getCodeMetadata(t, testContext.Accounts, scChild))
		assert.Equal(t, vmcommon.Ok, retCodeUpgrade)
		assert.Nil(t, errUpgrade)
	})
	t.Run("SC deploys child SC while payable by SC is active should handle the child SC upgrade", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		codeMetadata := vmcommon.CodeMetadata{
			Upgradeable: true,
		}
		codeMetadataHex := hex.EncodeToString(codeMetadata.ToBytes())
		scParent, scChild, retCodeDeploy, errDeploy := deployChildAndParentContracts(t, testContext, senderAddressBytes, codeMetadataHex, codeMetadataHex)
		assert.Equal(t, vmcommon.Ok, retCodeDeploy)
		assert.Nil(t, errDeploy)

		goodCodeMetadata := []byte{05, 02}
		t.Run("bad code metadata by value", func(t *testing.T) {
			badCodeMetadata := []byte{0xFF, 0xFF}
			retCode, errUpgrade := upgradeChildContract(t, testContext, senderAddressBytes, scParent, scChild, hex.EncodeToString(badCodeMetadata))

			assert.Equal(t, codeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, scChild))
			assert.Equal(t, vmcommon.ExecutionFailed, retCode)
			assert.Nil(t, errUpgrade)
		})
		t.Run("bad code metadata by number of values", func(t *testing.T) {
			badCodeMetadata := []byte{0xFF, 0xFF, 0xFF}
			retCode, errUpgrade := upgradeChildContract(t, testContext, senderAddressBytes, scParent, scChild, hex.EncodeToString(badCodeMetadata))

			assert.Equal(t, codeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, scChild))
			assert.Equal(t, vmcommon.ExecutionFailed, retCode)
			assert.Nil(t, errUpgrade)
		})
		t.Run("ok code metadata should work", func(t *testing.T) {
			retCode, errUpgrade := upgradeChildContract(t, testContext, senderAddressBytes, scParent, scChild, hex.EncodeToString(goodCodeMetadata))

			assert.Equal(t, goodCodeMetadata, getCodeMetadata(t, testContext.Accounts, scChild))
			assert.Equal(t, vmcommon.Ok, retCode)
			assert.Nil(t, errUpgrade)
		})
	})
}

func upgradeDummySC(tb testing.TB, testContext *vm.VMTestContext, senderAddressBytes []byte, scAddressBytes []byte, codeMetadataHex string) (vmcommon.ReturnCode, error) {
	transferOnCalls := big.NewInt(50)

	accnt, err := testContext.Accounts.LoadAccount(senderAddressBytes)
	require.Nil(tb, err)

	nonce := accnt.GetNonce()
	scCode := wasm.GetSCCode("../../testdata/counter/output/counter.wasm")
	tx := vm.CreateTx(
		senderAddressBytes,
		scAddressBytes,
		nonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		fmt.Sprintf("upgradeContract@%s@%s", scCode, codeMetadataHex),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil || returnCode != vmcommon.Ok {
		return returnCode, err
	}

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	return vmcommon.Ok, nil
}

func upgradeChildContract(
	tb testing.TB,
	testContext *vm.VMTestContext,
	senderAddressBytes []byte,
	scParentAddressBytes []byte,
	scChildAddressBytes []byte,
	codeMetadataHex string,
) (vmcommon.ReturnCode, error) {
	transferOnCalls := big.NewInt(50)

	childScCode := wasm.GetSCCode("../../testdata/hello-v1/output/answer.wasm")
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
		fmt.Sprintf("%s@%s@%s@%s", "upgradeChild", hex.EncodeToString(scChildAddressBytes), childScCode, codeMetadataHex),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	if err != nil || returnCode != vmcommon.Ok {
		return returnCode, err
	}
	require.Nil(tb, err)
	require.Equal(tb, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	return vmcommon.Ok, nil
}
