package resolvers

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ dataRetriever.Resolver = (*TxResolver)(nil)

// maxBuffToSendBulkTransactions represents max buffer size to send in bytes
const maxBuffToSendBulkTransactions = 1 << 18 // 256KB

// maxBuffToSendBulkMiniblocks represents max buffer size to send in bytes
const maxBuffToSendBulkMiniblocks = 1 << 18 // 256KB

// ArgTxResolver is the argument structure used to create new TxResolver instance
type ArgTxResolver struct {
	ArgBaseResolver
	TxPool            dataRetriever.ShardedDataCacherNotifier
	TxStorage         storage.Storer
	DataPacker        dataRetriever.DataPacker
	IsFullHistoryNode bool
}

// TxResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type TxResolver struct {
	*baseResolver
	messageProcessor
	baseStorageResolver
	txPool     dataRetriever.ShardedDataCacherNotifier
	dataPacker dataRetriever.DataPacker
}

// NewTxResolver creates a new transaction resolver
func NewTxResolver(arg ArgTxResolver) (*TxResolver, error) {
	err := checkArgTxResolver(arg)
	if err != nil {
		return nil, err
	}

	txResolver := &TxResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: arg.SenderResolver,
		},
		txPool:              arg.TxPool,
		baseStorageResolver: createBaseStorageResolver(arg.TxStorage, arg.IsFullHistoryNode),
		dataPacker:          arg.DataPacker,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshaller,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.RequestTopic(),
			throttler:        arg.Throttler,
		},
	}

	return txResolver, nil
}

func checkArgTxResolver(arg ArgTxResolver) error {
	err := checkArgBase(arg.ArgBaseResolver)
	if err != nil {
		return err
	}
	if check.IfNil(arg.TxPool) {
		return dataRetriever.ErrNilTxDataPool
	}
	if check.IfNil(arg.TxStorage) {
		return dataRetriever.ErrNilTxStorage
	}
	if check.IfNil(arg.DataPacker) {
		return dataRetriever.ErrNilDataPacker
	}
	return nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (txRes *TxResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	err := txRes.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	txRes.throttler.StartProcessing()
	defer txRes.throttler.EndProcessing()

	rd, err := txRes.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.HashType:
		err = txRes.resolveTxRequestByHash(rd.Value, message.Peer(), rd.Epoch)
	case dataRetriever.HashArrayType:
		err = txRes.resolveTxRequestByHashArray(rd.Value, message.Peer(), rd.Epoch)
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}

	if err != nil {
		err = fmt.Errorf("%w for hash %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

func (txRes *TxResolver) resolveTxRequestByHash(hash []byte, pid core.PeerID, epoch uint32) error {
	// TODO this can be optimized by searching in corresponding datapool (taken by topic name)
	tx, err := txRes.fetchTxAsByteSlice(hash, epoch)
	if err != nil {
		return err
	}

	b := &batch.Batch{
		Data: [][]byte{tx},
	}
	buff, err := txRes.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return txRes.Send(buff, pid)
}

func (txRes *TxResolver) fetchTxAsByteSlice(hash []byte, epoch uint32) ([]byte, error) {
	value, ok := txRes.txPool.SearchFirstData(hash)
	if ok {
		return txRes.marshalizer.Marshal(value)
	}

	buff, err := txRes.getFromStorage(hash, epoch)
	if err != nil {
		txRes.DebugHandler().LogFailedToResolveData(
			txRes.topic,
			hash,
			err,
		)

		return nil, err
	}

	txRes.DebugHandler().LogSucceededToResolveData(txRes.topic, hash)

	return buff, nil
}

func (txRes *TxResolver) resolveTxRequestByHashArray(hashesBuff []byte, pid core.PeerID, epoch uint32) error {
	// TODO this can be optimized by searching in corresponding datapool (taken by topic name)
	b := batch.Batch{}
	err := txRes.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	var errFetch error
	errorsFound := 0
	txsBuffSlice := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		tx, errTemp := txRes.fetchTxAsByteSlice(hash, epoch)
		if errTemp != nil {
			errFetch = fmt.Errorf("%w for hash %s", errTemp, logger.DisplayByteSlice(hash))
			// it might happen to error on a tx (maybe it is missing) but should continue
			// as to send back as many as it can
			log.Trace("fetchTxAsByteSlice missing",
				"error", errFetch.Error(),
				"hash", hash)
			errorsFound++

			continue
		}
		txsBuffSlice = append(txsBuffSlice, tx)
	}

	buffsToSend, errPack := txRes.dataPacker.PackDataInChunks(txsBuffSlice, maxBuffToSendBulkTransactions)
	if errPack != nil {
		return errPack
	}

	for _, buff := range buffsToSend {
		errSend := txRes.Send(buff, pid)
		if errSend != nil {
			return errSend
		}
	}

	if errFetch != nil {
		errFetch = fmt.Errorf("resolveTxRequestByHashArray last error %w from %d encountered errors", errFetch, errorsFound)
	}

	return errFetch
}

// IsInterfaceNil returns true if there is no value under the interface
func (txRes *TxResolver) IsInterfaceNil() bool {
	return txRes == nil
}
