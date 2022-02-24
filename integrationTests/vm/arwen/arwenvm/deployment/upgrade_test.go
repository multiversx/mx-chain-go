package deployment

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gasPrice = uint64(1)
const gasLimit = uint64(10000000)

func TestScUpgradeShouldManageCorrectlyTheCodeMetadata(t *testing.T) {
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
		upgradeDummySCReturningContractAddress(t, testContext, senderAddressBytes, contractAddress, hex.EncodeToString(newCodeMetadata))

		assert.Equal(t, newCodeMetadata, getCodeMetadata(t, testContext.Accounts, contractAddress))
	})
	t.Run("payable by SC is active, code metadata is parsed", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})

		contractAddress := deployDummySCReturningContractAddress(t, testContext, senderAddressBytes, "0502")

		deployCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: false,
			Upgradeable: true,
			Readable:    true,
		}
		assert.Equal(t, deployCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))

		newCodeMetadata := "FFFF"
		upgradeDummySCReturningContractAddress(t, testContext, senderAddressBytes, contractAddress, newCodeMetadata)

		upgradeCodeMetadata := vmcommon.CodeMetadata{
			Payable:     true,
			PayableBySC: true,
			Upgradeable: true,
			Readable:    true,
		}
		assert.Equal(t, upgradeCodeMetadata.ToBytes(), getCodeMetadata(t, testContext.Accounts, contractAddress))
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
		upgradeDummySCReturningContractAddress(t, testContext, senderAddressBytes, contractAddress, hex.EncodeToString(newCodeMetadata))
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
}

func upgradeDummySCReturningContractAddress(tb testing.TB, testContext *vm.VMTestContext, senderAddressBytes []byte, scAddressBytes []byte, codeMetadataHex string) {
	transferOnCalls := big.NewInt(50)

	accnt, err := testContext.Accounts.LoadAccount(senderAddressBytes)
	require.Nil(tb, err)

	nonce := accnt.GetNonce()
	scCode := arwen.GetSCCode("../../testdata/counter/output/counter.wasm")
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
	require.Nil(tb, err)
	require.Equal(tb, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)
}
