package settlement

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	apiCore "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
)

const (
	refundScrDeployGasLimit = uint64(100_000_000)
	refundScrCallGasLimit   = uint64(2_000_000)
	refundScrGasPrice       = uint64(1_000_000_000)
	refundScrOneEGLD        = int64(1_000_000_000_000_000_000)
	refundScrMaxNumOfBlocks = 20
)

// TestChainSimulator_RefundScrStatusFetchedByHash checks the status reported for the smart contract
// result of an in-shard refund, when it is fetched by hash through the transaction API (the call used
// by the proxy and by the explorers).
//
// With async execution (Supernova) enabled, an in-shard refund is settled during the parent transaction
// processing, so its SCR record is only held within the block's receipts and never becomes part of the
// block's execution results. Such a settled record must be reported as "success" and not as
// "not-executable-in-block".
func TestChainSimulator_RefundScrStatusFetchedByHash(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	initialBalance := big.NewInt(0).Mul(big.NewInt(10), big.NewInt(refundScrOneEGLD))
	owner, err := simulator.GenerateAndMintWalletAddress(shardID, initialBalance)
	require.NoError(t, err)

	// generate one block so the minting takes effect
	require.NoError(t, simulator.GenerateBlocks(1))

	facade := simulator.GetNodeHandler(shardID).GetFacadeHandler()

	// deploy the contract, using a huge gas limit; the deploy itself does not produce a refund SCR
	scCode := wasm.GetSCCode("../relayedTx/testData/adder.wasm")
	deployTxData := strings.Join([]string{scCode, wasm.VMTypeHex, wasm.DummyCodeMetadataHex, "00"}, "@")
	deployTx := &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(0),
		SndAddr:   owner.Bytes,
		RcvAddr:   make([]byte, 32),
		Data:      []byte(deployTxData),
		GasLimit:  refundScrDeployGasLimit,
		GasPrice:  refundScrGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	}

	resultDeploy, err := simulator.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, refundScrMaxNumOfBlocks)
	require.NoError(t, err)
	require.Equal(t, transaction.TxStatusSuccess, resultDeploy.Status)
	require.NotEmpty(t, resultDeploy.Logs.Events)

	scAddressBytes, err := simulator.GetNodeHandler(shardID).GetCoreComponents().AddressPubKeyConverter().Decode(resultDeploy.Logs.Events[0].Address)
	require.NoError(t, err)

	// call the contract with a gas limit larger than needed, so the owner receives an in-shard refund
	callTxData := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
	callTx := &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(0),
		SndAddr:   owner.Bytes,
		RcvAddr:   scAddressBytes,
		Data:      []byte(callTxData),
		GasLimit:  refundScrCallGasLimit,
		GasPrice:  refundScrGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	}

	resultCall, err := simulator.SendTxAndGenerateBlockTilTxIsExecuted(callTx, refundScrMaxNumOfBlocks)
	require.NoError(t, err)
	require.Equal(t, transaction.TxStatusSuccess, resultCall.Status)

	refundScrHash := findRefundScrHash(resultCall)
	require.NotEmpty(t, refundScrHash, "the in-shard refund SCR was not found on the executed transaction")

	// let the receipts of the block be saved/indexed
	require.NoError(t, simulator.GenerateBlocks(2))

	// the same SCR is reported as success within the block body, as it comes from the receipts storage
	blockWithRefund, err := facade.GetBlockByHash(resultCall.BlockHash, apiCore.BlockQueryOptions{WithTransactions: true})
	require.NoError(t, err)
	require.Equal(t, transaction.TxStatusSuccess, findTxStatusInBlock(blockWithRefund, refundScrHash))

	// fetch the refund SCR by hash through the transaction API
	refundScr, err := facade.GetTransaction(refundScrHash, true)
	require.NoError(t, err)
	require.Equal(t, transaction.TxStatusSuccess, refundScr.Status)
}

func findRefundScrHash(tx *transaction.ApiTransactionResult) string {
	for _, scr := range tx.SmartContractResults {
		if scr.IsRefund {
			return scr.Hash
		}
	}

	return ""
}

func findTxStatusInBlock(block *apiCore.Block, txHash string) transaction.TxStatus {
	for _, miniBlock := range block.MiniBlocks {
		for _, tx := range miniBlock.Transactions {
			if tx.Hash == txHash {
				return tx.Status
			}
		}
	}

	return ""
}
