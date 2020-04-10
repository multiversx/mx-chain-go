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
// TODO extract common functionality between this and transactionResolver
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
		err = mbRes.resolveMbRequestByHash(rd.Value, message.Peer())
	case dataRetriever.HashArrayType:
		err = mbRes.resolveMbRequestByHashArray(rd.Value, message.Peer())
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}

	if err != nil {
		err = fmt.Errorf("%w for hash %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

func (mbRes *miniblockResolver) resolveMbRequestByHash(hash []byte, pid p2p.PeerID) error {
	mb, err := mbRes.fetchMbAsByteSlice(hash)
	if err != nil {
		return err
	}

	buffToSend, err := mbRes.marshalizer.Marshal(&batch.Batch{Data: [][]byte{mb}})
	if err != nil {
		return err
	}

	return mbRes.Send(buffToSend, pid)
}

func (mbRes *miniblockResolver) fetchMbAsByteSlice(hash []byte) ([]byte, error) {
	value, ok := mbRes.miniBlockPool.Peek(hash)
	if ok {
		return mbRes.marshalizer.Marshal(value)
	}

	return mbRes.miniBlockStorage.Get(hash)
}

func (mbRes *miniblockResolver) resolveMbRequestByHashArray(mbBuff []byte, pid p2p.PeerID) error {
	b := batch.Batch{}
	err := mbRes.marshalizer.Unmarshal(&b, mbBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	var errFetch error
	errorsFound := 0
	mbsBuffSlice := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		mb, errTemp := mbRes.fetchMbAsByteSlice(hash)
		if errTemp != nil {
			errFetch = errTemp
			log.Trace("fetchMbAsByteSlice missing",
				"error", errFetch.Error(),
				"hash", hash)
			errorsFound++

			continue
		}
		mbsBuffSlice = append(mbsBuffSlice, mb)
	}

	buffsToSend, errPack := mbRes.dataPacker.PackDataInChunks(mbsBuffSlice, maxBuffToSendBulkMiniblocks)
	if errPack != nil {
		return errPack
	}

	for _, buff := range buffsToSend {
		errSend := mbRes.Send(buff, pid)
		if errSend != nil {
			return errSend
		}
	}

	if errFetch != nil {
		errFetch = fmt.Errorf("resolveMbRequestByHashArray last error %w from %d encountered errors", errFetch, errorsFound)
	}

	return errFetch
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
	b := &batch.Batch{
		Data: hashes,
	}
	hash, err := mbRes.marshalizer.Marshal(b)

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
