package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// maxBuffToSendBulkTransactions represents max buffer size to send in bytes
var maxBuffToSendBulkTransactions = 2 << 17 //128KB

// TxResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type TxResolver struct {
	dataRetriever.TopicResolverSender
	txPool      dataRetriever.ShardedDataCacherNotifier
	txStorage   storage.Storer
	marshalizer marshal.Marshalizer
	dataPacker  dataRetriever.DataPacker
}

// NewTxResolver creates a new transaction resolver
func NewTxResolver(
	senderResolver dataRetriever.TopicResolverSender,
	txPool dataRetriever.ShardedDataCacherNotifier,
	txStorage storage.Storer,
	marshalizer marshal.Marshalizer,
	dataPacker dataRetriever.DataPacker,
) (*TxResolver, error) {

	if senderResolver == nil || senderResolver.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if txPool == nil || txPool.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if txStorage == nil || txStorage.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilTxStorage
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if dataPacker == nil || dataPacker.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilDataPacker
	}

	txResolver := &TxResolver{
		TopicResolverSender: senderResolver,
		txPool:              txPool,
		txStorage:           txStorage,
		marshalizer:         marshalizer,
		dataPacker:          dataPacker,
	}

	return txResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (txRes *TxResolver) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	rd := &dataRetriever.RequestData{}
	err := rd.Unmarshal(txRes.marshalizer, message)
	if err != nil {
		return err
	}

	if rd.Value == nil {
		return dataRetriever.ErrNilValue
	}

	switch rd.Type {
	case dataRetriever.HashType:
		buff, err := txRes.resolveTxRequestByHash(rd.Value)
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

// IsInterfaceNil returns true if there is no value under the interface
func (txRes *TxResolver) IsInterfaceNil() bool {
	if txRes == nil {
		return true
	}
	return false
}
