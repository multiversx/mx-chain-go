package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// maxBuffToSendBulkTransactions represents max buffer size to send in bytes
var maxBuffToSendBulkTransactions = 2 << 17 //128KB

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
			topic:            arg.SenderResolver.Topic(),
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
		var buff []byte
		buff, err = txRes.resolveTxRequestByHash(rd.Value)
		if err != nil {
			return err
		}
		return txRes.Send(buff, message.Peer())
	case dataRetriever.HashArrayType:
		return txRes.resolveTxRequestByHashArray(rd.Value, message.Peer())
	default:
		return dataRetriever.ErrRequestTypeNotImplemented
	}
}

func (txRes *TxResolver) resolveTxRequestByHash(hash []byte) ([]byte, error) {
	//TODO this can be optimized by searching in corresponding datapool (taken by topic name)
	txsBuff := make([][]byte, 0)

	tx, err := txRes.fetchTxAsByteSlice(hash)
	if err != nil {
		return nil, err
	}

	txsBuff = append(txsBuff, tx)
	buffToSend, err := txRes.marshalizer.Marshal(txsBuff)
	if err != nil {
		return nil, err
	}

	return buffToSend, nil
}

func (txRes *TxResolver) fetchTxAsByteSlice(hash []byte) ([]byte, error) {
	value, ok := txRes.txPool.SearchFirstData(hash)
	if ok {
		txBuff, err := txRes.marshalizer.Marshal(value)
		if err != nil {
			return nil, err
		}
		return txBuff, nil
	}

	return txRes.txStorage.SearchFirst(hash)
}

func (txRes *TxResolver) resolveTxRequestByHashArray(hashesBuff []byte, pid p2p.PeerID) error {
	//TODO this can be optimized by searching in corresponding datapool (taken by topic name)
	hashes := make([][]byte, 0)
	err := txRes.marshalizer.Unmarshal(&hashes, hashesBuff)
	if err != nil {
		return err
	}

	txsBuffSlice := make([][]byte, 0)
	for _, hash := range hashes {
		var tx []byte
		tx, err = txRes.fetchTxAsByteSlice(hash)
		if err != nil {
			//it might happen to error on a tx (maybe it is missing) but should continue
			// as to send back as many as it can
			log.Trace("fetchTxAsByteSlice missing",
				"error", err.Error(),
				"hash", hash)
			continue
		}
		txsBuffSlice = append(txsBuffSlice, tx)
	}

	buffsToSend, err := txRes.dataPacker.PackDataInChunks(txsBuffSlice, maxBuffToSendBulkTransactions)

	for _, buff := range buffsToSend {
		err = txRes.Send(buff, pid)
		if err != nil {
			return err
		}
	}

	return nil
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
	buffHashes, err := txRes.marshalizer.Marshal(hashes)
	if err != nil {
		return err
	}

	return txRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashArrayType,
		Value: buffHashes,
		Epoch: epoch,
	})
}

// IsInterfaceNil returns true if there is no value under the interface
func (txRes *TxResolver) IsInterfaceNil() bool {
	return txRes == nil
}
