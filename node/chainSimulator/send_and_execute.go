package chainSimulator

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

func (s *simulator) sendTx(tx *transaction.Transaction) (string, error) {
	shardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(tx.SndAddr)
	err := s.GetNodeHandler(shardID).GetFacadeHandler().ValidateTransaction(tx)
	if err != nil {
		return "", err
	}

	node := s.GetNodeHandler(shardID)
	txHash, err := core.CalculateHash(node.GetCoreComponents().InternalMarshalizer(), node.GetCoreComponents().Hasher(), tx)
	if err != nil {
		return "", err
	}

	txHashHex := hex.EncodeToString(txHash)
	_, err = node.GetFacadeHandler().SendBulkTransactions([]*transaction.Transaction{tx})
	if err != nil {
		return "", err
	}

	log.Info("############## send transaction ##############", "txHash", txHashHex)

	return txHashHex, nil
}

func (s *simulator) SendTxsAndGenerateBlockTilTxIsExecuted(txsToSend []*transaction.Transaction, maxNumOfBlockToGenerateWhenExecutingTx int) ([]*transaction.ApiTransactionResult, error) {
	hashTxIndex := make(map[string]int)
	for idx, txToSend := range txsToSend {
		txHashHex, err := s.sendTx(txToSend)
		if err != nil {
			return nil, err
		}

		hashTxIndex[txHashHex] = idx
	}

	time.Sleep(100 * time.Millisecond)

	txsFromAPI := make([]*transaction.ApiTransactionResult, 3)
	for count := 0; count < maxNumOfBlockToGenerateWhenExecutingTx; count++ {
		err := s.GenerateBlocks(1)
		if err != nil {
			return nil, err
		}

		for txHash := range hashTxIndex {
			destinationShardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(txsToSend[hashTxIndex[txHash]].RcvAddr)
			tx, errGet := s.GetNodeHandler(destinationShardID).GetFacadeHandler().GetTransaction(txHash, true)
			if errGet == nil && tx.Status != transaction.TxStatusPending {
				log.Info("############## transaction was executed ##############", "txHash", txHash)

				txsFromAPI[hashTxIndex[txHash]] = tx
				delete(hashTxIndex, txHash)
				continue
			}
		}
		if len(hashTxIndex) == 0 {
			return txsFromAPI, nil
		}
	}

	return nil, errors.New("something went wrong transactions are still in pending")
}
