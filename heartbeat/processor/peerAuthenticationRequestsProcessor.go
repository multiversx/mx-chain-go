package processor

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("heartbeat/processor")

const (
	minMessagesInChunk      = 1
	minDelayBetweenRequests = time.Second
	minTimeout              = time.Second
	minMessagesThreshold    = 0.5
)

// ArgPeerAuthenticationRequestsProcessor represents the arguments for the peer authentication request processor
type ArgPeerAuthenticationRequestsProcessor struct {
	RequestHandler         process.RequestHandler
	NodesCoordinator       heartbeat.NodesCoordinator
	PeerAuthenticationPool storage.Cacher
	ShardId                uint32
	Epoch                  uint32
	MessagesInChunk        uint32
	MinPeersThreshold      float32
	DelayBetweenRequests   time.Duration
	MaxTimeout             time.Duration
}

// PeerAuthenticationRequestsProcessor defines the component that sends the requests for peer authentication messages
type PeerAuthenticationRequestsProcessor struct {
	requestHandler         process.RequestHandler
	nodesCoordinator       heartbeat.NodesCoordinator
	peerAuthenticationPool storage.Cacher
	shardId                uint32
	epoch                  uint32
	messagesInChunk        uint32
	minPeersThreshold      float32
	delayBetweenRequests   time.Duration
	maxTimeout             time.Duration
	cancel                 func()
}

// NewPeerAuthenticationRequestsProcessor creates a new instance of PeerAuthenticationRequestsProcessor
func NewPeerAuthenticationRequestsProcessor(args ArgPeerAuthenticationRequestsProcessor) (*PeerAuthenticationRequestsProcessor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	processor := &PeerAuthenticationRequestsProcessor{
		requestHandler:         args.RequestHandler,
		nodesCoordinator:       args.NodesCoordinator,
		peerAuthenticationPool: args.PeerAuthenticationPool,
		shardId:                args.ShardId,
		epoch:                  args.Epoch,
		messagesInChunk:        args.MessagesInChunk,
		minPeersThreshold:      args.MinPeersThreshold,
		delayBetweenRequests:   args.DelayBetweenRequests,
		maxTimeout:             args.MaxTimeout,
	}

	var ctx context.Context
	ctx, processor.cancel = context.WithCancel(context.Background())

	go processor.startRequestingMessages(ctx)

	return processor, nil
}

func checkArgs(args ArgPeerAuthenticationRequestsProcessor) error {
	if check.IfNil(args.RequestHandler) {
		return heartbeat.ErrNilRequestHandler
	}
	if check.IfNil(args.NodesCoordinator) {
		return heartbeat.ErrNilNodesCoordinator
	}
	if check.IfNil(args.PeerAuthenticationPool) {
		return heartbeat.ErrNilPeerAuthenticationPool
	}
	if args.MessagesInChunk < minMessagesInChunk {
		return fmt.Errorf("%w for MessagesInChunk, provided %d, min expected %d",
			heartbeat.ErrInvalidValue, args.MessagesInChunk, minMessagesInChunk)
	}
	if args.MinPeersThreshold < minMessagesThreshold {
		return fmt.Errorf("%w for MinPeersThreshold, provided %f, min expected %f",
			heartbeat.ErrInvalidValue, args.MinPeersThreshold, minMessagesThreshold)
	}
	if args.DelayBetweenRequests < minDelayBetweenRequests {
		return fmt.Errorf("%w for DelayBetweenRequests, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.DelayBetweenRequests, minDelayBetweenRequests)
	}
	if args.MaxTimeout < minTimeout {
		return fmt.Errorf("%w for MaxTimeout, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.MaxTimeout, minTimeout)
	}

	return nil
}

func (processor *PeerAuthenticationRequestsProcessor) startRequestingMessages(ctx context.Context) {
	defer processor.cancel()

	sortedValidatorsKeys, err := processor.getSortedValidatorsKeys()
	if err != nil {
		return
	}

	// first request messages by chunks
	processor.requestKeysChunks(sortedValidatorsKeys)

	// start endless loop until enough messages received or timeout reached
	requestsTimer := time.NewTimer(processor.delayBetweenRequests)
	timeoutTimer := time.NewTimer(processor.maxTimeout)
	for {
		if processor.isThresholdReached(sortedValidatorsKeys) {
			log.Debug("received enough messages, closing PeerAuthenticationRequestsProcessor go routine")
			return
		}

		requestsTimer.Reset(processor.delayBetweenRequests)
		select {
		case <-requestsTimer.C:
			processor.requestMissingKeys(sortedValidatorsKeys)
		case <-timeoutTimer.C:
			log.Debug("timeout reached, not enough messages received, closing PeerAuthenticationRequestsProcessor go routine")
			return
		case <-ctx.Done():
			log.Debug("closing PeerAuthenticationRequestsProcessor go routine")
			return
		}
	}
}

func (processor *PeerAuthenticationRequestsProcessor) requestKeysChunks(keys [][]byte) {
	maxChunks := processor.getMaxChunks(keys)
	for chunkIndex := uint32(0); chunkIndex < maxChunks; chunkIndex++ {
		processor.requestHandler.RequestPeerAuthenticationsChunk(processor.shardId, chunkIndex)

		time.Sleep(processor.delayBetweenRequests)
	}
}

func (processor *PeerAuthenticationRequestsProcessor) getSortedValidatorsKeys() ([][]byte, error) {
	validatorsPKsMap, err := processor.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(processor.epoch)
	if err != nil {
		return nil, err
	}

	validatorsPKs := make([][]byte, 0)
	for _, shardValidators := range validatorsPKsMap {
		validatorsPKs = append(validatorsPKs, shardValidators...)
	}

	sort.Slice(validatorsPKs, func(i, j int) bool {
		return bytes.Compare(validatorsPKs[i], validatorsPKs[j]) < 0
	})

	return validatorsPKs, nil
}

func (processor *PeerAuthenticationRequestsProcessor) getMaxChunks(dataBuff [][]byte) uint32 {
	maxChunks := len(dataBuff) / int(processor.messagesInChunk)
	if len(dataBuff)%int(processor.messagesInChunk) != 0 {
		maxChunks++
	}

	return uint32(maxChunks)
}

func (processor *PeerAuthenticationRequestsProcessor) isThresholdReached(sortedValidatorsKeys [][]byte) bool {
	minKeysExpected := float32(len(sortedValidatorsKeys)) * processor.minPeersThreshold
	keysInCache := processor.peerAuthenticationPool.Keys()

	return float32(len(keysInCache)) >= minKeysExpected
}

func (processor *PeerAuthenticationRequestsProcessor) requestMissingKeys(sortedValidatorsKeys [][]byte) {
	missingKeys := processor.getMissingKeys(sortedValidatorsKeys)
	if len(missingKeys) == 0 {
		return
	}

	processor.requestHandler.RequestPeerAuthenticationsByHashes(processor.shardId, missingKeys)
}

func (processor *PeerAuthenticationRequestsProcessor) getMissingKeys(sortedValidatorsKeys [][]byte) [][]byte {
	validatorsMap := make(map[string]bool, len(sortedValidatorsKeys))
	for _, key := range sortedValidatorsKeys {
		validatorsMap[string(key)] = false
	}

	keysInCache := processor.peerAuthenticationPool.Keys()
	for _, key := range keysInCache {
		validatorsMap[string(key)] = true
	}

	missingKeys := make([][]byte, 0)
	for mKey, mVal := range validatorsMap {
		if mVal {
			missingKeys = append(missingKeys, []byte(mKey))
		}
	}

	return missingKeys
}

// Close closes the internal components
func (processor *PeerAuthenticationRequestsProcessor) Close() error {
	if processor.cancel != nil {
		log.Debug("closing PeerAuthenticationRequestsProcessor go routine")
		processor.cancel()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (processor *PeerAuthenticationRequestsProcessor) IsInterfaceNil() bool {
	return processor == nil
}
