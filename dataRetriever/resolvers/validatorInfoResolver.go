package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const minNumOfValidatorInfo = 5

// ArgValidatorInfoResolver is the argument structure used to create a new validator info resolver instance
type ArgValidatorInfoResolver struct {
	SenderResolver                  dataRetriever.TopicResolverSender
	Marshaller                      marshal.Marshalizer
	AntifloodHandler                dataRetriever.P2PAntifloodHandler
	Throttler                       dataRetriever.ResolverThrottler
	ValidatorInfoPool               storage.Cacher
	ValidatorInfoStorage            storage.Storer
	IsFullHistoryNode               bool
	MaxNumOfValidatorInfoInResponse int
}

// validatorInfoResolver is a wrapper over Resolver that is specialized in resolving validator info requests
type validatorInfoResolver struct {
	dataRetriever.TopicResolverSender
	messageProcessor
	baseStorageResolver
	validatorInfoPool               storage.Cacher
	validatorInfoStorage            storage.Storer
	maxNumOfValidatorInfoInResponse int
}

// NewValidatorInfoResolver creates a validator info resolver
func NewValidatorInfoResolver(args ArgValidatorInfoResolver) (*validatorInfoResolver, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &validatorInfoResolver{
		TopicResolverSender: args.SenderResolver,
		messageProcessor: messageProcessor{
			marshalizer:      args.Marshaller,
			antifloodHandler: args.AntifloodHandler,
			throttler:        args.Throttler,
			topic:            args.SenderResolver.RequestTopic(),
		},
		baseStorageResolver:             createBaseStorageResolver(args.ValidatorInfoStorage, args.IsFullHistoryNode),
		validatorInfoPool:               args.ValidatorInfoPool,
		validatorInfoStorage:            args.ValidatorInfoStorage,
		maxNumOfValidatorInfoInResponse: args.MaxNumOfValidatorInfoInResponse,
	}, nil
}

func checkArgs(args ArgValidatorInfoResolver) error {
	if check.IfNil(args.SenderResolver) {
		return dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(args.Marshaller) {
		return dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.AntifloodHandler) {
		return dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(args.Throttler) {
		return dataRetriever.ErrNilThrottler
	}
	if check.IfNil(args.ValidatorInfoPool) {
		return dataRetriever.ErrNilValidatorInfoPool
	}
	if check.IfNil(args.ValidatorInfoStorage) {
		return dataRetriever.ErrNilValidatorInfoStorage
	}
	if args.MaxNumOfValidatorInfoInResponse < minNumOfValidatorInfo {
		return dataRetriever.ErrInvalidNumOfValidatorInfo
	}

	return nil
}

// RequestDataFromHash requests validator info from other peers by hash
func (res *validatorInfoResolver) RequestDataFromHash(hash []byte, epoch uint32) error {
	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
			Epoch: epoch,
		},
		[][]byte{hash},
	)
}

// RequestDataFromHashArray requests validator info from other peers by hash array
func (res *validatorInfoResolver) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := res.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
			Epoch: epoch,
		},
		hashes,
	)
}

// ProcessReceivedMessage represents the callback func from the p2p.Messenger that is called each time a new message is received
// (for the topic this validator was registered to, usually a request topic)
func (res *validatorInfoResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	err := res.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	res.throttler.StartProcessing()
	defer res.throttler.EndProcessing()

	rd, err := res.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.HashType:
		return res.resolveHashRequest(rd.Value, rd.Epoch, fromConnectedPeer)
	case dataRetriever.HashArrayType:
		return res.resolveMultipleHashesRequest(rd.Value, rd.Epoch, fromConnectedPeer)
	}

	return fmt.Errorf("%w for value %s", dataRetriever.ErrRequestTypeNotImplemented, logger.DisplayByteSlice(rd.Value))
}

// resolveHashRequest sends the response for a hash request
func (res *validatorInfoResolver) resolveHashRequest(hash []byte, epoch uint32, pid core.PeerID) error {
	data, err := res.fetchValidatorInfoByteSlice(hash, epoch)
	if err != nil {
		return err
	}

	return res.marshalAndSend([][]byte{data}, pid)
}

// resolveMultipleHashesRequest sends the response for a hash array type request
func (res *validatorInfoResolver) resolveMultipleHashesRequest(hashesBuff []byte, epoch uint32, pid core.PeerID) error {
	b := batch.Batch{}
	err := res.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	validatorInfoForHashes, err := res.fetchValidatorInfoForHashes(hashes, epoch)
	if err != nil {
		return fmt.Errorf("resolveMultipleHashesRequest error %w from buff %s", err, hashesBuff)
	}

	return res.sendValidatorInfoForHashes(validatorInfoForHashes, pid)
}

func (res *validatorInfoResolver) sendValidatorInfoForHashes(validatorInfoForHashes [][]byte, pid core.PeerID) error {
	if len(validatorInfoForHashes) > res.maxNumOfValidatorInfoInResponse {
		return res.sendLargeDataBuff(validatorInfoForHashes, pid)
	}

	return res.marshalAndSend(validatorInfoForHashes, pid)
}

func (res *validatorInfoResolver) sendLargeDataBuff(dataBuff [][]byte, pid core.PeerID) error {
	chunksMap := res.splitDataBuffIntoChunks(dataBuff)
	for _, chunk := range chunksMap {
		err := res.marshalAndSend(chunk, pid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (res *validatorInfoResolver) splitDataBuffIntoChunks(dataBuff [][]byte) map[int][][]byte {
	chunksMap := make(map[int][][]byte)
	currentChunk := make([][]byte, 0)
	currentChunkSize := 0
	chunkIndex := 0
	for _, data := range dataBuff {
		if currentChunkSize == res.maxNumOfValidatorInfoInResponse {
			chunksMap[chunkIndex] = currentChunk
			chunkIndex++
			currentChunk = make([][]byte, 0)
			currentChunkSize = 0
		}

		currentChunk = append(currentChunk, data)
		currentChunkSize++
	}
	chunksMap[chunkIndex] = currentChunk

	return chunksMap
}

func (res *validatorInfoResolver) fetchValidatorInfoForHashes(hashes [][]byte, epoch uint32) ([][]byte, error) {
	validatorInfos := make([][]byte, 0)
	for _, hash := range hashes {
		validatorInfoForHash, _ := res.fetchValidatorInfoByteSlice(hash, epoch)
		if validatorInfoForHash != nil {
			validatorInfos = append(validatorInfos, validatorInfoForHash)
		}
	}

	if len(validatorInfos) == 0 {
		return nil, dataRetriever.ErrValidatorInfoNotFound
	}

	return validatorInfos, nil
}

func (res *validatorInfoResolver) fetchValidatorInfoByteSlice(hash []byte, epoch uint32) ([]byte, error) {
	data, ok := res.validatorInfoPool.Get(hash)
	if ok {
		return res.marshalizer.Marshal(data)
	}

	buff, err := res.getFromStorage(hash, epoch)
	if err != nil {
		res.ResolverDebugHandler().LogFailedToResolveData(
			res.topic,
			hash,
			err,
		)
		return nil, err
	}

	res.ResolverDebugHandler().LogSucceededToResolveData(res.topic, hash)

	return buff, nil
}

func (res *validatorInfoResolver) marshalAndSend(data [][]byte, pid core.PeerID) error {
	b := &batch.Batch{
		Data: data,
	}
	buff, err := res.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return res.Send(buff, pid)
}

// SetResolverDebugHandler sets a resolver debug handler
func (res *validatorInfoResolver) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	return res.TopicResolverSender.SetResolverDebugHandler(handler)
}

// SetNumPeersToQuery sets the number of intra shard and cross shard peers to query
func (res *validatorInfoResolver) SetNumPeersToQuery(intra int, cross int) {
	res.TopicResolverSender.SetNumPeersToQuery(intra, cross)
}

// NumPeersToQuery returns the number of intra shard and cross shard peers to query
func (res *validatorInfoResolver) NumPeersToQuery() (int, int) {
	return res.TopicResolverSender.NumPeersToQuery()
}

// Close returns nil
func (res *validatorInfoResolver) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *validatorInfoResolver) IsInterfaceNil() bool {
	return res == nil
}
