package helpers

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationTests/chainSimulator/helpers")

func computeTxHash(chainSimulator ChainSimulator, tx *transaction.Transaction) (string, error) {
	txBytes, err := chainSimulator.GetNodeHandler(1).GetCoreComponents().InternalMarshalizer().Marshal(tx)
	if err != nil {
		return "", err
	}

	txHasBytes := chainSimulator.GetNodeHandler(1).GetCoreComponents().Hasher().Compute(string(txBytes))
	return hex.EncodeToString(txHasBytes), nil
}

// SendTxAndGenerateBlockTilTxIsExecuted will the provided transaction and generate block
func SendTxAndGenerateBlockTilTxIsExecuted(
	t *testing.T,
	chainSimulator ChainSimulator,
	txToSend *transaction.Transaction,
	maxNumOfBlockToGenerateWhenExecutingTx int,
) *transaction.ApiTransactionResult {
	shardID := chainSimulator.GetNodeHandler(0).GetShardCoordinator().ComputeId(txToSend.SndAddr)
	err := chainSimulator.GetNodeHandler(shardID).GetFacadeHandler().ValidateTransaction(txToSend)
	require.Nil(t, err)

	txHash, err := computeTxHash(chainSimulator, txToSend)
	require.Nil(t, err)
	log.Info("############## send transaction ##############", "txHash", txHash)

	_, err = chainSimulator.GetNodeHandler(shardID).GetFacadeHandler().SendBulkTransactions([]*transaction.Transaction{txToSend})
	require.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	destinationShardID := chainSimulator.GetNodeHandler(0).GetShardCoordinator().ComputeId(txToSend.RcvAddr)
	for count := 0; count < maxNumOfBlockToGenerateWhenExecutingTx; count++ {
		err = chainSimulator.GenerateBlocks(1)
		require.Nil(t, err)

		tx, errGet := chainSimulator.GetNodeHandler(destinationShardID).GetFacadeHandler().GetTransaction(txHash, true)
		if errGet == nil && tx.Status != transaction.TxStatusPending {
			log.Info("############## transaction was executed ##############", "txHash", txHash)
			return tx
		}
	}

	t.Error("something went wrong transaction is still in pending")
	t.FailNow()

	return nil
}

// AddValidatorKeysInMultiKey will add provided keys in the multi key handler
func AddValidatorKeysInMultiKey(t *testing.T, chainSimulator ChainSimulator, keysBase64 []string) [][]byte {
	privateKeysHex := make([]string, 0, len(keysBase64))
	for _, keyBase64 := range keysBase64 {
		privateKeyHex, err := base64.StdEncoding.DecodeString(keyBase64)
		require.Nil(t, err)

		privateKeysHex = append(privateKeysHex, string(privateKeyHex))
	}

	privateKeysBytes := make([][]byte, 0, len(privateKeysHex))
	for _, keyHex := range privateKeysHex {
		privateKeyBytes, err := hex.DecodeString(keyHex)
		require.Nil(t, err)

		privateKeysBytes = append(privateKeysBytes, privateKeyBytes)
	}

	err := chainSimulator.AddValidatorKeys(privateKeysBytes)
	require.Nil(t, err)

	return privateKeysBytes
}

// GenerateBlsPrivateKeys will generate bls keys
func GenerateBlsPrivateKeys(t *testing.T, numOfKeys int) ([][]byte, []string) {
	blockSigningGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	secretKeysBytes := make([][]byte, 0, numOfKeys)
	blsKeysHex := make([]string, 0, numOfKeys)
	for idx := 0; idx < numOfKeys; idx++ {
		secretKey, publicKey := blockSigningGenerator.GeneratePair()

		secretKeyBytes, err := secretKey.ToByteArray()
		require.Nil(t, err)

		secretKeysBytes = append(secretKeysBytes, secretKeyBytes)

		publicKeyBytes, err := publicKey.ToByteArray()
		require.Nil(t, err)

		blsKeysHex = append(blsKeysHex, hex.EncodeToString(publicKeyBytes))
	}

	return secretKeysBytes, blsKeysHex
}
