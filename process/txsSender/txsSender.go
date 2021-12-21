package txsSender

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("txsSender")

type NetworkMessenger interface {
	io.Closer
	// BroadcastOnChannelBlocking asynchronously waits until it can send a
	// message on the channel, but once it is able to, it synchronously sends the
	// message, blocking until sending is completed.
	BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

type txsSender struct {
	marshaller       marshal.Marshalizer
	shardCoordinator storage.ShardCoordinator
	networkMessenger NetworkMessenger

	txSentCounter            uint32
	txAcumulator             core.Accumulator
	currentSendingGoRoutines int32
}

// SendBulkTransactions sends the provided transactions as a bulk, optimizing transfer between nodes
func (ts *txsSender) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	if len(txs) == 0 {
		return 0, node.ErrNoTxToProcess
	}

	ts.addTransactionsToSendPipe(txs)

	return uint64(len(txs)), nil
}

func (ts *txsSender) addTransactionsToSendPipe(txs []*transaction.Transaction) {
	if check.IfNil(ts.txAcumulator) {
		log.Error("node has a nil tx accumulator instance")
		return
	}

	for _, tx := range txs {
		ts.txAcumulator.AddData(tx)
	}
}

func (ts *txsSender) sendFromTxAccumulator(ctx context.Context) {
	outputChannel := ts.txAcumulator.OutputChannel()

	for {
		select {
		case objs := <-outputChannel:
			{
				if len(objs) == 0 {
					break
				}

				txs := make([]*transaction.Transaction, 0, len(objs))
				for _, obj := range objs {
					tx, ok := obj.(*transaction.Transaction)
					if !ok {
						continue
					}

					txs = append(txs, tx)
				}

				atomic.AddUint32(&ts.txSentCounter, uint32(len(txs)))

				ts.sendBulkTransactions(txs)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ts *txsSender) sendBulkTransactions(txs []*transaction.Transaction) {
	transactionsByShards := make(map[uint32][][]byte)
	log.Trace("txsSender.sendBulkTransactions sending txs",
		"num", len(txs),
	)

	for _, tx := range txs {
		senderShardId := ts.shardCoordinator.ComputeId(tx.SndAddr)

		marshalizedTx, err := ts.marshaller.Marshal(tx)
		if err != nil {
			log.Warn("txsSender.sendBulkTransactions",
				"marshalizer error", err,
			)
			continue
		}

		transactionsByShards[senderShardId] = append(transactionsByShards[senderShardId], marshalizedTx)
	}

	numOfSentTxs := uint64(0)
	for shardId, txsForShard := range transactionsByShards {
		err := ts.sendBulkTransactionsFromShard(txsForShard, shardId)
		if err != nil {
			log.Debug("sendBulkTransactionsFromShard", "error", err.Error())
		} else {
			numOfSentTxs += uint64(len(txsForShard))
		}
	}
}

func (ts *txsSender) sendBulkTransactionsFromShard(transactions [][]byte, senderShardId uint32) error {
	dataPacker, err := partitioning.NewSimpleDataPacker(ts.marshaller)
	if err != nil {
		return err
	}

	// the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + ts.shardCoordinator.CommunicationIdentifier(senderShardId)

	packets, err := dataPacker.PackDataInChunks(transactions, common.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	atomic.AddInt32(&ts.currentSendingGoRoutines, int32(len(packets)))
	for _, buff := range packets {
		go func(bufferToSend []byte) {
			log.Trace("txsSender.sendBulkTransactionsFromShard",
				"topic", identifier,
				"size", len(bufferToSend),
			)
			err = ts.networkMessenger.BroadcastOnChannelBlocking(
				node.SendTransactionsPipe,
				identifier,
				bufferToSend,
			)
			if err != nil {
				log.Debug("txsSender.BroadcastOnChannelBlocking", "error", err.Error())
			}

			atomic.AddInt32(&ts.currentSendingGoRoutines, -1)
		}(buff)
	}

	return nil
}
