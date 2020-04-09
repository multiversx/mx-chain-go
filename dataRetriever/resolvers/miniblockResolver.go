package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgMiniblockResolver is the argument structure used to create a new miniblockResolver instance
type ArgMiniblockResolver struct {
	SenderResolver   dataRetriever.TopicResolverSender
	MiniBlockPool    storage.Cacher
	MiniBlockStorage storage.Storer
	Marshalizer      marshal.Marshalizer
	AntifloodHandler dataRetriever.P2PAntifloodHandler
	Throttler        dataRetriever.ResolverThrottler
	DataPacker       dataRetriever.DataPacker
}

// miniblockResolver is a wrapper over Resolver that is specialized in resolving miniblocks requests
type miniblockResolver struct {
	dataRetriever.TopicResolverSender
	messageProcessor
	miniBlockPool    storage.Cacher
	miniBlockStorage storage.Storer
	dataPacker       dataRetriever.DataPacker
}

// NewMiniblockResolver creates a miniblock resolver
func NewMiniblockResolver(arg ArgMiniblockResolver) (*miniblockResolver, error) {
	if check.IfNil(arg.SenderResolver) {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(arg.MiniBlockPool) {
		return nil, dataRetriever.ErrNilMiniblocksPool
	}
	if check.IfNil(arg.MiniBlockStorage) {
		return nil, dataRetriever.ErrNilMiniblocksStorage
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.AntifloodHandler) {
		return nil, dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.Throttler) {
		return nil, dataRetriever.ErrNilThrottler
	}
	if check.IfNil(arg.DataPacker) {
		return nil, dataRetriever.ErrNilDataPacker
	}

	mbResolver := &miniblockResolver{
		TopicResolverSender: arg.SenderResolver,
		miniBlockPool:       arg.MiniBlockPool,
		miniBlockStorage:    arg.MiniBlockStorage,
		dataPacker:          arg.DataPacker,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshalizer,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.RequestTopic(),
			throttler:        arg.Throttler,
		},
	}

	return mbResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (mbRes *miniblockResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	err := mbRes.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	mbRes.throttler.StartProcessing()
	defer mbRes.throttler.EndProcessing()

	rd, err := mbRes.parseReceivedMessage(message)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.HashType:
		var buff []byte
		buff, err = mbRes.resolveMbRequestByHash(rd.Value)
		if err != nil {
			return err
		}
		return mbRes.Send(buff, message.Peer())
	case dataRetriever.HashArrayType:
		return mbRes.resolveMbRequestByHashArray(rd.Value, message.Peer())
	default:
		return dataRetriever.ErrRequestTypeNotImplemented
	}
}

func (mbRes *miniblockResolver) resolveMbRequestByHash(hash []byte) ([]byte, error) {
	mb, err := mbRes.fetchMbAsByteSlice(hash)
	if err != nil {
		return nil, err
	}

	buffToSend, err := mbRes.marshalizer.Marshal(&batch.Batch{Data: [][]byte{mb}})
	if err != nil {
		return nil, err
	}

	return buffToSend, nil
}

func (mbRes *miniblockResolver) fetchMbAsByteSlice(hash []byte) ([]byte, error) {
	value, ok := mbRes.miniBlockPool.Peek(hash)
	if ok {
		txBuff, err := mbRes.marshalizer.Marshal(value)
		if err != nil {
			return nil, err
		}
		return txBuff, nil
	}

	return mbRes.miniBlockStorage.SearchFirst(hash)
}

func (mbRes *miniblockResolver) resolveMbRequestByHashArray(mbBuff []byte, pid p2p.PeerID) error {
	b := batch.Batch{}
	err := mbRes.marshalizer.Unmarshal(&b, mbBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	mbsBuffSlice := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		mb, errFetch := mbRes.fetchMbAsByteSlice(hash)
		if errFetch != nil {
			err = errFetch
			log.Trace("fetchTxAsByteSlice missing",
				"error", err.Error(),
				"hash", hash)
			continue
		}
		mbsBuffSlice = append(mbsBuffSlice, mb)
	}

	buffsToSend, errPack := mbRes.dataPacker.PackDataInChunks(mbsBuffSlice, maxBuffToSendBulkTransactions)
	if errPack != nil {
		return errPack
	}

	for _, buff := range buffsToSend {
		errSend := mbRes.Send(buff, pid)
		if errSend != nil {
			return errSend
		}
	}

	return err
}

// RequestDataFromHash requests a block body from other peers having input the block body hash
func (mbRes *miniblockResolver) RequestDataFromHash(hash []byte, epoch uint32) error {
	return mbRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
		Epoch: epoch,
	})
}

// RequestDataFromHashArray requests a block body from other peers having input the block body hash
func (mbRes *miniblockResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
	hash, err := mbRes.marshalizer.Marshal(&batch.Batch{Data: hashes})

	if err != nil {
		return err
	}

	return mbRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashArrayType,
		Value: hash,
	})
}

// SetNumPeersToQuery will set the number of intra shard and cross shard number of peer to query
func (mbRes *miniblockResolver) SetNumPeersToQuery(intra int, cross int) {
	mbRes.TopicResolverSender.SetNumPeersToQuery(intra, cross)
}

// GetNumPeersToQuery will return the number of intra shard and cross shard number of peer to query
func (mbRes *miniblockResolver) GetNumPeersToQuery() (int, int) {
	return mbRes.TopicResolverSender.GetNumPeersToQuery()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbRes *miniblockResolver) IsInterfaceNil() bool {
	return mbRes == nil
}
