package resolvers

import (
	"fmt"

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

// ArgEquivalentProofsResolver is the argument structure used to create a new equivalent proofs resolver instance
type ArgEquivalentProofsResolver struct {
	ArgBaseResolver
	DataPacker              dataRetriever.DataPacker
	EquivalentProofsStorage storage.Storer
	EquivalentProofsPool    processor.EquivalentProofsPool
	IsFullHistoryNode       bool
}

type equivalentProofsResolver struct {
	*baseResolver
	baseStorageResolver
	messageProcessor
	dataPacker              dataRetriever.DataPacker
	equivalentProofsStorage storage.Storer
	equivalentProofsPool    processor.EquivalentProofsPool
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
		dataPacker:              args.DataPacker,
		equivalentProofsStorage: args.EquivalentProofsStorage,
		equivalentProofsPool:    args.EquivalentProofsPool,
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
		return dataRetriever.ErrNilProofsStorage
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
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}
	if err != nil {
		return nil, fmt.Errorf("%w for value %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return []byte{}, nil
}

// resolveHashRequest sends the response for a hash request
func (res *equivalentProofsResolver) resolveHashRequest(hash []byte, epoch uint32, pid core.PeerID, source p2p.MessageHandler) error {
	data, err := res.fetchEquivalentProofAsByteSlice(hash, epoch)
	if err != nil {
		return err
	}

	return res.marshalAndSend(data, pid, source)
}

// resolveMultipleHashesRequest sends the response for multiple hashes request
func (res *equivalentProofsResolver) resolveMultipleHashesRequest(hashesBuff []byte, epoch uint32, pid core.PeerID, source p2p.MessageHandler) error {
	b := batch.Batch{}
	err := res.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	equivalentProofsForHashes, err := res.fetchEquivalentProofsSlicesForHeaders(hashes, epoch)
	if err != nil {
		return fmt.Errorf("resolveMultipleHashesRequest error %w from buff %x", err, hashesBuff)
	}

	return res.sendEquivalentProofsForHashes(equivalentProofsForHashes, pid, source)
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

func (res *equivalentProofsResolver) marshalAndSend(data []byte, pid core.PeerID, source p2p.MessageHandler) error {
	b := &batch.Batch{
		Data: [][]byte{data},
	}
	buff, err := res.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return res.Send(buff, pid, source)
}

// fetchEquivalentProofsSlicesForHeaders fetches all equivalent proofs for the given header hashes
func (res *equivalentProofsResolver) fetchEquivalentProofsSlicesForHeaders(headerHashes [][]byte, epoch uint32) ([][]byte, error) {
	equivalentProofs := make([][]byte, 0)
	for _, headerHash := range headerHashes {
		equivalentProofForHash, _ := res.fetchEquivalentProofAsByteSlice(headerHash, epoch)
		if equivalentProofForHash != nil {
			equivalentProofs = append(equivalentProofs, equivalentProofForHash)
		}
	}

	if len(equivalentProofs) == 0 {
		return nil, dataRetriever.ErrEquivalentProofsNotFound
	}

	return equivalentProofs, nil
}

// fetchEquivalentProofAsByteSlice returns the value from equivalent proofs storage if exists
func (res *equivalentProofsResolver) fetchEquivalentProofAsByteSlice(headerHash []byte, epoch uint32) ([]byte, error) {
	proof, err := res.equivalentProofsPool.GetProofByHash(headerHash)
	if err != nil {
		return res.getFromStorage(headerHash, epoch)
	}

	return res.marshalizer.Marshal(proof)
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *equivalentProofsResolver) IsInterfaceNil() bool {
	return res == nil
}
