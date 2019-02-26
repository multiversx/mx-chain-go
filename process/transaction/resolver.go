package transaction

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TxResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type TxResolver struct {
	process.TopicResolverSender
	txPool      data.ShardedDataCacherNotifier
	txStorage   storage.Storer
	marshalizer marshal.Marshalizer
}

// NewTxResolver creates a new transaction resolver
func NewTxResolver(
	senderResolver process.TopicResolverSender,
	txPool data.ShardedDataCacherNotifier,
	txStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*TxResolver, error) {

	if senderResolver == nil {
		return nil, process.ErrNilResolverSender
	}

	if txPool == nil {
		return nil, process.ErrNilTxDataPool
	}

	if txStorage == nil {
		return nil, process.ErrNilTxStorage
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
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
	rd := &process.RequestData{}
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

func (txRes *TxResolver) resolveTxRequest(rd *process.RequestData) ([]byte, error) {
	if rd.Type != process.HashType {
		return nil, process.ErrResolveNotHashType
	}

	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	value, ok := txRes.txPool.SearchFirstData(rd.Value)
	if !ok {
		return txRes.txStorage.Get(rd.Value)
	}

	buff, err := txRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RequestDataFromHash requests a transaction from other peers having input the tx hash
func (txRes *TxResolver) RequestDataFromHash(hash []byte) error {
	return txRes.SendOnRequestTopic(&process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
