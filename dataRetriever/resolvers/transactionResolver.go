package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TxResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type TxResolver struct {
	dataRetriever.TopicResolverSender
	txPool      data.ShardedDataCacherNotifier
	txStorage   storage.Storer
	marshalizer marshal.Marshalizer
}

// NewTxResolver creates a new transaction resolver
func NewTxResolver(
	senderResolver dataRetriever.TopicResolverSender,
	txPool data.ShardedDataCacherNotifier,
	txStorage storage.Storer,
	marshalizer marshal.Marshalizer,
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

	txResolver := &TxResolver{
		TopicResolverSender: senderResolver,
		txPool:              txPool,
		txStorage:           txStorage,
		marshalizer:         marshalizer,
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

	buff, err := txRes.resolveTxRequest(rd)
	if err != nil {
		return err
	}

	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil
	}

	return txRes.Send(buff, message.Peer())
}

func (txRes *TxResolver) resolveTxRequest(rd *dataRetriever.RequestData) ([]byte, error) {
	//TODO - implement other types such as HashArrayType for an array of transaction
	// This should be made in future subtasks belonging to EN-1520 story
	if rd.Type != dataRetriever.HashType {
		return nil, dataRetriever.ErrResolveNotHashType
	}

	if rd.Value == nil {
		return nil, dataRetriever.ErrNilValue
	}

	//TODO this can be optimized by searching in corresponding datapool (taken by topic name)
	txsBuff := make([][]byte, 0)
	value, ok := txRes.txPool.SearchFirstData(rd.Value)
	if ok {
		txBuff, err := txRes.marshalizer.Marshal(value)
		if err != nil {
			return nil, err
		}
		txsBuff = append(txsBuff, txBuff)
	} else {
		buff, err := txRes.txStorage.Get(rd.Value)
		if err != nil {
			return nil, err
		}
		txsBuff = append(txsBuff, buff)
	}

	buffToSend, err := txRes.marshalizer.Marshal(txsBuff)
	if err != nil {
		return nil, err
	}

	return buffToSend, nil
}

// RequestDataFromHash requests a transaction from other peers having input the tx hash
func (txRes *TxResolver) RequestDataFromHash(hash []byte) error {
	return txRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
	})
}
