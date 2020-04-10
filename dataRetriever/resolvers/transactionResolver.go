package resolvers

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// maxBuffToSendBulkTransactions represents max buffer size to send in bytes
const maxBuffToSendBulkTransactions = 1 << 18 //256KB

// maxBuffToSendBulkMiniblocks represents max buffer size to send in bytes
const maxBuffToSendBulkMiniblocks = 1 << 18 //256KB

// ArgTxResolver is the argument structure used to create new TxResolver instance
type ArgTxResolver struct {
	SenderResolver   dataRetriever.TopicResolverSender
	TxPool           dataRetriever.ShardedDataCacherNotifier
	TxStorage        storage.Storer
	Marshalizer      marshal.Marshalizer
	DataPacker       dataRetriever.DataPacker
	AntifloodHandler dataRetriever.P2PAntifloodHandler
	Throttler        dataRetriever.ResolverThrottler
}

// TxResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type TxResolver struct {
	dataRetriever.TopicResolverSender
	messageProcessor
	txPool     dataRetriever.ShardedDataCacherNotifier
	txStorage  storage.Storer
	dataPacker dataRetriever.DataPacker
}

// NewTxResolver creates a new transaction resolver
func NewTxResolver(arg ArgTxResolver) (*TxResolver, error) {
	if check.IfNil(arg.SenderResolver) {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(arg.TxPool) {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if check.IfNil(arg.TxStorage) {
		return nil, dataRetriever.ErrNilTxStorage
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.DataPacker) {
		return nil, dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(arg.AntifloodHandler) {
		return nil, dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.Throttler) {
		return nil, dataRetriever.ErrNilThrottler
	}

	txResolver := &TxResolver{
		TopicResolverSender: arg.SenderResolver,
		txPool:              arg.TxPool,
		txStorage:           arg.TxStorage,
		dataPacker:          arg.DataPacker,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshalizer,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.RequestTopic(),
			throttler:        arg.Throttler,
		},
	}

	return txResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (txRes *TxResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	err := txRes.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	txRes.throttler.StartProcessing()
	defer txRes.throttler.EndProcessing()

	rd, err := txRes.parseReceivedMessage(message)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.HashType:
		err = txRes.resolveTxRequestByHash(rd.Value, message.Peer())
	case dataRetriever.HashArrayType:
		err = txRes.resolveTxRequestByHashArray(rd.Value, message.Peer())
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}

	if err != nil {
		err = fmt.Errorf("%w for hash %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

func (txRes *TxResolver) resolveTxRequestByHash(hash []byte, pid p2p.PeerID) error {
	//TODO this can be optimized by searching in corresponding datapool (taken by topic name)
	tx, err := txRes.fetchTxAsByteSlice(hash)
	if err != nil {
		return err
	}

	buff, err := txRes.marshalizer.Marshal(&batch.Batch{Data: [][]byte{tx}})
	if err != nil {
		return err
	}

	return txRes.Send(buff, pid)
}

func (txRes *TxResolver) fetchTxAsByteSlice(hash []byte) ([]byte, error) {
	value, ok := txRes.txPool.SearchFirstData(hash)
	if ok {
		return txRes.marshalizer.Marshal(value)
	}

	return txRes.txStorage.SearchFirst(hash)
}

func (txRes *TxResolver) resolveTxRequestByHashArray(hashesBuff []byte, pid p2p.PeerID) error {
	//TODO this can be optimized by searching in corresponding datapool (taken by topic name)
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
		tx, errTemp := txRes.fetchTxAsByteSlice(hash)
		if errTemp != nil {
			errFetch = errTemp
			err = fmt.Errorf("%w for hash %s", errFetch, logger.DisplayByteSlice(hash))
			//it might happen to error on a tx (maybe it is missing) but should continue
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

// RequestDataFromHash requests a transaction from other peers having input the tx hash
func (txRes *TxResolver) RequestDataFromHash(hash []byte, epoch uint32) error {
	return txRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
		Epoch: epoch,
	})
}

// RequestDataFromHashArray requests a list of tx hashes from other peers
func (txRes *TxResolver) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	buffHashes, err := txRes.marshalizer.Marshal(&batch.Batch{Data: hashes})
	if err != nil {
		return err
	}

	return txRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashArrayType,
		Value: buffHashes,
		Epoch: epoch,
	})
}

// SetNumPeersToQuery will set the number of intra shard and cross shard number of peer to query
func (txRes *TxResolver) SetNumPeersToQuery(intra int, cross int) {
	txRes.TopicResolverSender.SetNumPeersToQuery(intra, cross)
}

// GetNumPeersToQuery will return the number of intra shard and cross shard number of peer to query
func (txRes *TxResolver) GetNumPeersToQuery() (int, int) {
	return txRes.TopicResolverSender.GetNumPeersToQuery()
}

// IsInterfaceNil returns true if there is no value under the interface
func (txRes *TxResolver) IsInterfaceNil() bool {
	return txRes == nil
}
