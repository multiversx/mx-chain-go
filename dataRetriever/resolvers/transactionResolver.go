package resolvers

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// MaxBuffToSendBulkTransactions represents max buffer size to send in bytes
var MaxBuffToSendBulkTransactions = 2 << 17 //128KB

// TxResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type TxResolver struct {
	dataRetriever.TopicResolverSender
	txPool        dataRetriever.ShardedDataCacherNotifier
	txStorage     storage.Storer
	marshalizer   marshal.Marshalizer
	sliceSplitter dataRetriever.SliceSplitter
}

// NewTxResolver creates a new transaction resolver
func NewTxResolver(
	senderResolver dataRetriever.TopicResolverSender,
	txPool dataRetriever.ShardedDataCacherNotifier,
	txStorage storage.Storer,
	marshalizer marshal.Marshalizer,
	sliceSplitter dataRetriever.SliceSplitter,
) (*TxResolver, error) {

	if senderResolver == nil {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if txPool == nil {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if txStorage == nil {
		return nil, dataRetriever.ErrNilTxStorage
	}
	if marshalizer == nil {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if sliceSplitter == nil {
		return nil, dataRetriever.ErrNilSliceSplitter
	}

	txResolver := &TxResolver{
		TopicResolverSender: senderResolver,
		txPool:              txPool,
		txStorage:           txStorage,
		marshalizer:         marshalizer,
		sliceSplitter:       sliceSplitter,
	}

	return txResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (txRes *TxResolver) ProcessReceivedMessage(message p2p.MessageP2P) error {
	rd := &dataRetriever.RequestData{}
	err := rd.Unmarshal(txRes.marshalizer, message)
	if err != nil {
		return err
	}

	if rd.Value == nil {
		return dataRetriever.ErrNilValue
	}

	if rd.Type == dataRetriever.HashType {
		buff, err := txRes.resolveTxRequestByHash(rd.Value)
		if err != nil {
			return err
		}
		return txRes.Send(buff, message.Peer())
	}
	if rd.Type == dataRetriever.HashArrayType {
		return txRes.resolveTxRequestByHashArray(rd.Value, message.Peer())
	}

	return dataRetriever.ErrRequestTypeNotImplemented
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

	return txRes.txStorage.Get(hash)
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
		tx, err := txRes.fetchTxAsByteSlice(hash)
		if err != nil {
			//it might happen to error on a tx (maybe it is missing) but should continue
			// as to send back as many as it can
			log.Debug(err.Error())
			continue
		}
		txsBuffSlice = append(txsBuffSlice, tx)
	}

	err = txRes.sliceSplitter.SendDataInChunks(
		txsBuffSlice,
		func(buff []byte) error {
			return txRes.Send(buff, pid)
		},
		MaxBuffToSendBulkTransactions,
	)

	return err
}

// RequestDataFromHash requests a transaction from other peers having input the tx hash
func (txRes *TxResolver) RequestDataFromHash(hash []byte) error {
	return txRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
	})
}

// RequestDataFromHashArray requests a list of tx hashes from other peers
func (txRes *TxResolver) RequestDataFromHashArray(hashes [][]byte) error {
	buffHashes, err := txRes.marshalizer.Marshal(hashes)
	if err != nil {
		return err
	}

	return txRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashArrayType,
		Value: buffHashes,
	})
}
