package components

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgsSyncedTxsSender is a holder struct for all necessary arguments to create a NewSyncedTxsSender
type ArgsSyncedTxsSender struct {
	Marshaller       marshal.Marshalizer
	ShardCoordinator sharding.Coordinator
	NetworkMessenger NetworkMessenger
	DataPacker       process.DataPacker
}

type syncedTxsSender struct {
	marshaller       marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	networkMessenger NetworkMessenger
	dataPacker       process.DataPacker
}

// NewSyncedTxsSender creates a new instance of syncedTxsSender
func NewSyncedTxsSender(args ArgsSyncedTxsSender) (*syncedTxsSender, error) {
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.NetworkMessenger) {
		return nil, process.ErrNilMessenger
	}
	if check.IfNil(args.DataPacker) {
		return nil, dataRetriever.ErrNilDataPacker
	}

	ret := &syncedTxsSender{
		marshaller:       args.Marshaller,
		shardCoordinator: args.ShardCoordinator,
		networkMessenger: args.NetworkMessenger,
		dataPacker:       args.DataPacker,
	}

	return ret, nil
}

// SendBulkTransactions sends the provided transactions as a bulk, optimizing transfer between nodes
func (sender *syncedTxsSender) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	if len(txs) == 0 {
		return 0, process.ErrNoTxToProcess
	}

	sender.sendBulkTransactions(txs)

	return uint64(len(txs)), nil
}

func (sender *syncedTxsSender) sendBulkTransactions(txs []*transaction.Transaction) {
	transactionsByShards := make(map[uint32][][]byte)
	for _, tx := range txs {
		marshalledTx, err := sender.marshaller.Marshal(tx)
		if err != nil {
			log.Warn("txsSender.sendBulkTransactions",
				"marshaller error", err,
			)
			continue
		}

		senderShardId := sender.shardCoordinator.ComputeId(tx.SndAddr)
		transactionsByShards[senderShardId] = append(transactionsByShards[senderShardId], marshalledTx)
	}

	for shardId, txsForShard := range transactionsByShards {
		err := sender.sendBulkTransactionsFromShard(txsForShard, shardId)
		log.LogIfError(err)
	}
}

func (sender *syncedTxsSender) sendBulkTransactionsFromShard(transactions [][]byte, senderShardId uint32) error {
	// the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + sender.shardCoordinator.CommunicationIdentifier(senderShardId)

	packets, err := sender.dataPacker.PackDataInChunks(transactions, common.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	for _, buff := range packets {
		sender.networkMessenger.Broadcast(identifier, buff)
	}

	return nil
}

// Close returns nil
func (sender *syncedTxsSender) Close() error {
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sender *syncedTxsSender) IsInterfaceNil() bool {
	return sender == nil
}
