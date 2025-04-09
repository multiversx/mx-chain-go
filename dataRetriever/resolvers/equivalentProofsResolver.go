package resolvers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// maxBuffToSendEquivalentProofs represents max buffer size to send in bytes
const maxBuffToSendEquivalentProofs = 1 << 18 // 256KB

const (
	keySeparator   = "-"
	expectedKeyLen = 2
	hashIndex      = 0
	shardIndex     = 1
	nonceIndex     = 0
)

// ArgEquivalentProofsResolver is the argument structure used to create a new equivalent proofs resolver instance
type ArgEquivalentProofsResolver struct {
	ArgBaseResolver
	DataPacker                       dataRetriever.DataPacker
	EquivalentProofsStorage          storage.Storer
	EquivalentProofsNonceHashStorage storage.Storer
	EquivalentProofsPool             processor.EquivalentProofsPool
	IsFullHistoryNode                bool
}

type equivalentProofsResolver struct {
	*baseResolver
	baseStorageResolver
	messageProcessor
	dataPacker                       dataRetriever.DataPacker
	equivalentProofsNonceHashStorage storage.Storer
	equivalentProofsPool             processor.EquivalentProofsPool
}

// NewEquivalentProofsResolver creates an equivalent proofs resolver
func NewEquivalentProofsResolver(args ArgEquivalentProofsResolver) (*equivalentProofsResolver, error) {
	err := checkArgEquivalentProofsResolver(args)
	if err != nil {
		return nil, err
	}

	return &equivalentProofsResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: args.SenderResolver,
		},
		baseStorageResolver: createBaseStorageResolver(args.EquivalentProofsStorage, args.IsFullHistoryNode),
		messageProcessor: messageProcessor{
			marshalizer:      args.Marshaller,
			antifloodHandler: args.AntifloodHandler,
			throttler:        args.Throttler,
			topic:            args.SenderResolver.RequestTopic(),
		},
		dataPacker:                       args.DataPacker,
		equivalentProofsNonceHashStorage: args.EquivalentProofsNonceHashStorage,
		equivalentProofsPool:             args.EquivalentProofsPool,
	}, nil
}

func checkArgEquivalentProofsResolver(args ArgEquivalentProofsResolver) error {
	err := checkArgBase(args.ArgBaseResolver)
	if err != nil {
		return err
	}
	if check.IfNil(args.DataPacker) {
		return dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(args.EquivalentProofsStorage) {
		return fmt.Errorf("%w for EquivalentProofsStorage", dataRetriever.ErrNilProofsStorage)
	}
	if check.IfNil(args.EquivalentProofsNonceHashStorage) {
		return fmt.Errorf("%w for EquivalentProofsNonceHashStorage", dataRetriever.ErrNilProofsStorage)
	}
	if check.IfNil(args.EquivalentProofsPool) {
		return dataRetriever.ErrNilProofsPool
	}

	return nil
}

// ProcessReceivedMessage represents the callback func from the p2p.Messenger that is called each time a new message is received
// (for the topic this validator was registered to, usually a request topic)
func (res *equivalentProofsResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, error) {
	err := res.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return nil, err
	}

	res.throttler.StartProcessing()
	defer res.throttler.EndProcessing()

	rd, err := res.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return nil, err
	}

	switch rd.Type {
	case dataRetriever.HashType:
		return nil, res.resolveHashRequest(rd.Value, rd.Epoch, message.Peer(), source)
	case dataRetriever.HashArrayType:
		return nil, res.resolveMultipleHashesRequest(rd.Value, rd.Epoch, message.Peer(), source)
	case dataRetriever.NonceType:
		return nil, res.resolveNonceRequest(rd.Value, rd.Epoch, message.Peer(), source)
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}
	if err != nil {
		return nil, fmt.Errorf("%w for value %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return []byte{}, nil
}

// resolveHashRequest sends the response for a hash request
func (res *equivalentProofsResolver) resolveHashRequest(hashShardKey []byte, epoch uint32, pid core.PeerID, source p2p.MessageHandler) error {
	headerHash, shardID, err := getHashAndShard(hashShardKey)
	if err != nil {
		return fmt.Errorf("resolveHashRequest.getHashAndShard error %w", err)
	}

	data, err := res.fetchEquivalentProofAsByteSlice(headerHash, shardID, epoch)
	if err != nil {
		return fmt.Errorf("resolveHashRequest.fetchEquivalentProofAsByteSlice error %w", err)
	}

	return res.Send(data, pid, source)
}

// resolveMultipleHashesRequest sends the response for multiple hashes request
func (res *equivalentProofsResolver) resolveMultipleHashesRequest(hashShardKeysBuff []byte, epoch uint32, pid core.PeerID, source p2p.MessageHandler) error {
	b := batch.Batch{}
	err := res.marshalizer.Unmarshal(&b, hashShardKeysBuff)
	if err != nil {
		return err
	}
	hashShardKeys := b.Data

	equivalentProofsForHashes, err := res.fetchEquivalentProofsSlicesForHeaders(hashShardKeys, epoch)
	if err != nil {
		return fmt.Errorf("resolveMultipleHashesRequest.fetchEquivalentProofsSlicesForHeaders error %w", err)
	}

	return res.sendEquivalentProofsForHashes(equivalentProofsForHashes, pid, source)
}

// resolveNonceRequest sends the response for a nonce request
func (res *equivalentProofsResolver) resolveNonceRequest(nonceShardKey []byte, epoch uint32, pid core.PeerID, source p2p.MessageHandler) error {
	data, err := res.fetchEquivalentProofFromNonceAsByteSlice(nonceShardKey, epoch)
	if err != nil {
		return fmt.Errorf("resolveNonceRequest.fetchEquivalentProofFromNonceAsByteSlice error %w", err)
	}

	return res.Send(data, pid, source)
}

// sendEquivalentProofsForHashes sends multiple equivalent proofs for specific hashes
func (res *equivalentProofsResolver) sendEquivalentProofsForHashes(dataBuff [][]byte, pid core.PeerID, source p2p.MessageHandler) error {
	buffsToSend, err := res.dataPacker.PackDataInChunks(dataBuff, maxBuffToSendEquivalentProofs)
	if err != nil {
		return err
	}

	for _, buff := range buffsToSend {
		err = res.Send(buff, pid, source)
		if err != nil {
			return err
		}
	}

	return nil
}

// fetchEquivalentProofsSlicesForHeaders fetches all equivalent proofs for the given header hashes
func (res *equivalentProofsResolver) fetchEquivalentProofsSlicesForHeaders(hashShardKeys [][]byte, epoch uint32) ([][]byte, error) {
	equivalentProofs := make([][]byte, 0)
	for _, hashShardKey := range hashShardKeys {
		headerHash, shardID, err := getHashAndShard(hashShardKey)
		if err != nil {
			return nil, err
		}

		equivalentProofForHash, _ := res.fetchEquivalentProofAsByteSlice(headerHash, shardID, epoch)
		if equivalentProofForHash != nil {
			equivalentProofs = append(equivalentProofs, equivalentProofForHash)
		}
	}

	if len(equivalentProofs) == 0 {
		return nil, dataRetriever.ErrEquivalentProofsNotFound
	}

	return equivalentProofs, nil
}

// fetchEquivalentProofAsByteSlice returns the value from equivalent proofs pool or storage if exists
func (res *equivalentProofsResolver) fetchEquivalentProofAsByteSlice(headerHash []byte, shardID uint32, epoch uint32) ([]byte, error) {
	proof, err := res.equivalentProofsPool.GetProof(shardID, headerHash)
	if err != nil {
		return res.getFromStorage(headerHash, epoch)
	}

	return res.marshalizer.Marshal(proof)
}

// fetchEquivalentProofFromNonceAsByteSlice returns the value from equivalent proofs pool or storage if exists
func (res *equivalentProofsResolver) fetchEquivalentProofFromNonceAsByteSlice(nonceShardKey []byte, epoch uint32) ([]byte, error) {
	headerNonce, shardID, err := getNonceAndShard(nonceShardKey)
	if err != nil {
		return nil, fmt.Errorf("fetchEquivalentProofFromNonceAsByteSlice.getNonceAndShard error %w", err)
	}

	proof, err := res.equivalentProofsPool.GetProofByNonce(headerNonce, shardID)
	if err != nil {
		return res.getProofFromStorageByNonce(nonceShardKey, epoch)
	}

	return res.marshalizer.Marshal(proof)
}

// getProofFromStorageByNonce returns the value from equivalent storage if exists
func (res *equivalentProofsResolver) getProofFromStorageByNonce(nonceShardKey []byte, epoch uint32) ([]byte, error) {
	headerHash, err := res.equivalentProofsNonceHashStorage.Get(nonceShardKey)
	if err != nil {
		return nil, err
	}

	return res.getFromStorage(headerHash, epoch)
}

func getHashAndShard(hashShardKey []byte) ([]byte, uint32, error) {
	hashShardKeyStr := string(hashShardKey)
	result := strings.Split(hashShardKeyStr, keySeparator)
	if len(result) != expectedKeyLen {
		return nil, 0, dataRetriever.ErrInvalidHashShardKey
	}

	hash := []byte(result[hashIndex])
	shard, err := strconv.Atoi(result[shardIndex])
	if err != nil {
		return nil, 0, err
	}

	return hash, uint32(shard), nil
}

func getNonceAndShard(nonceShardKey []byte) (uint64, uint32, error) {
	nonceShardKeyStr := string(nonceShardKey)
	result := strings.Split(nonceShardKeyStr, keySeparator)
	if len(result) != expectedKeyLen {
		return 0, 0, dataRetriever.ErrInvalidNonceShardKey
	}

	nonce, err := strconv.Atoi(result[nonceIndex])
	if err != nil {
		return 0, 0, err
	}

	shard, err := strconv.Atoi(result[shardIndex])
	if err != nil {
		return 0, 0, err
	}

	return uint64(nonce), uint32(shard), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *equivalentProofsResolver) IsInterfaceNil() bool {
	return res == nil
}
