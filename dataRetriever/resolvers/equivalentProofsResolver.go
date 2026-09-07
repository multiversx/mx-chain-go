package resolvers

import (
	"fmt"
	"runtime/debug"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/storage"
)

// maxBuffToSendEquivalentProofs represents max buffer size to send in bytes
const maxBuffToSendEquivalentProofs = 1 << 18 // 256KB

// ArgEquivalentProofsResolver is the argument structure used to create a new equivalent proofs resolver instance
type ArgEquivalentProofsResolver struct {
	ArgBaseResolver
	DataPacker           dataRetriever.DataPacker
	Storage              dataRetriever.StorageService
	EquivalentProofsPool processor.EquivalentProofsPool
	NonceConverter       typeConverters.Uint64ByteSliceConverter
	IsFullHistoryNode    bool
}

type equivalentProofsResolver struct {
	*baseResolver
	baseStorageResolver
	messageProcessor
	dataPacker           dataRetriever.DataPacker
	storage              dataRetriever.StorageService
	equivalentProofsPool processor.EquivalentProofsPool
	nonceConverter       typeConverters.Uint64ByteSliceConverter
}

// NewEquivalentProofsResolver creates an equivalent proofs resolver
func NewEquivalentProofsResolver(args ArgEquivalentProofsResolver) (*equivalentProofsResolver, error) {
	err := checkArgEquivalentProofsResolver(args)
	if err != nil {
		return nil, err
	}

	equivalentProofsStorage, err := args.Storage.GetStorer(dataRetriever.ProofsUnit)
	if err != nil {
		return nil, err
	}

	return &equivalentProofsResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: args.SenderResolver,
		},
		baseStorageResolver: createBaseStorageResolver(equivalentProofsStorage, args.IsFullHistoryNode),
		messageProcessor: messageProcessor{
			marshalizer:      args.Marshaller,
			antifloodHandler: args.AntifloodHandler,
			throttler:        args.Throttler,
			topic:            args.SenderResolver.RequestTopic(),
		},
		dataPacker:           args.DataPacker,
		storage:              args.Storage,
		equivalentProofsPool: args.EquivalentProofsPool,
		nonceConverter:       args.NonceConverter,
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
	if check.IfNil(args.Storage) {
		return dataRetriever.ErrNilStore
	}
	if check.IfNil(args.EquivalentProofsPool) {
		return dataRetriever.ErrNilProofsPool
	}
	if check.IfNil(args.NonceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}

	return nil
}

// ProcessReceivedMessage represents the callback func from the p2p.Messenger that is called each time a new message is received
// (for the topic this validator was registered to, usually a request topic)
func (res *equivalentProofsResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) (msg []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			logTrieNodes.Error("panic recovered", "peer", fromConnectedPeer, "panic", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("panic in equivalentProofsResolver.ProcessReceivedMessage: %v", r)
		}
	}()

	err = res.canProcessMessage(message, fromConnectedPeer)
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
	headerHash, shardID, err := common.GetHashAndShardFromKey(hashShardKey)
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
	hashShardKeys := deduplicateHashes(b.Data)

	equivalentProofsForHashes, err := res.fetchEquivalentProofsSlicesForHeaders(hashShardKeys, epoch)
	if err != nil {
		return fmt.Errorf("resolveMultipleHashesRequest.fetchEquivalentProofsSlicesForHeaders error %w", err)
	}

	return res.sendEquivalentProofsForHashes(equivalentProofsForHashes, pid, source)
}

// resolveNonceRequest sends the response for a nonce request; all proofs held at the nonce are
// sent, each as a separate message (the interceptor expects one proof per message)
func (res *equivalentProofsResolver) resolveNonceRequest(nonceShardKey []byte, epoch uint32, pid core.PeerID, source p2p.MessageHandler) error {
	proofsBuffs, err := res.fetchEquivalentProofsFromNonceAsByteSlices(nonceShardKey, epoch)
	if err != nil {
		return fmt.Errorf("resolveNonceRequest.fetchEquivalentProofsFromNonceAsByteSlices error %w", err)
	}

	for _, buff := range proofsBuffs {
		err = res.Send(buff, pid, source)
		if err != nil {
			return err
		}
	}

	return nil
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
		headerHash, shardID, err := common.GetHashAndShardFromKey(hashShardKey)
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

// fetchEquivalentProofsFromNonceAsByteSlices returns all proofs held at the nonce from the
// equivalent proofs pool, or the single stored one from storage if the pool has none
func (res *equivalentProofsResolver) fetchEquivalentProofsFromNonceAsByteSlices(nonceShardKey []byte, epoch uint32) ([][]byte, error) {
	headerNonce, shardID, err := common.GetNonceAndShardFromKey(nonceShardKey)
	if err != nil {
		return nil, fmt.Errorf("fetchEquivalentProofsFromNonceAsByteSlices.getNonceAndShard error %w", err)
	}

	proofs, err := res.equivalentProofsPool.GetProofsByNonce(headerNonce, shardID)
	if err != nil {
		buff, errStorage := res.getProofFromStorageByNonce(headerNonce, shardID, epoch)
		if errStorage != nil {
			return nil, errStorage
		}

		return [][]byte{buff}, nil
	}

	proofsBuffs := make([][]byte, 0, len(proofs))
	for _, proof := range proofs {
		buff, errMarshal := res.marshalizer.Marshal(proof)
		if errMarshal != nil {
			return nil, errMarshal
		}

		proofsBuffs = append(proofsBuffs, buff)
	}

	return proofsBuffs, nil
}

// getProofFromStorageByNonce returns the value from equivalent storage if exists
func (res *equivalentProofsResolver) getProofFromStorageByNonce(headerNonce uint64, shardID uint32, epoch uint32) ([]byte, error) {
	storer, err := res.getStorerForShard(shardID)
	if err != nil {
		return nil, err
	}

	nonceBytes := res.nonceConverter.ToByteSlice(headerNonce)
	headerHash, err := storer.SearchFirst(nonceBytes)
	if err != nil {
		return nil, err
	}

	return res.getFromStorage(headerHash, epoch)
}

func (res *equivalentProofsResolver) getStorerForShard(shardID uint32) (storage.Storer, error) {
	if shardID == core.MetachainShardId {
		return res.storage.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	}

	return res.storage.GetStorer(dataRetriever.GetHdrNonceHashDataUnit(shardID))
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *equivalentProofsResolver) IsInterfaceNil() bool {
	return res == nil
}
