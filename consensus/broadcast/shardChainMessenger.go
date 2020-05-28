package broadcast

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var _ consensus.BroadcastMessenger = (*shardChainMessenger)(nil)

type delayedBroadcastData struct {
	headerHash   []byte
	miniblocks   map[uint32][]byte
	transactions map[string][][]byte
}

type shardChainMessenger struct {
	*commonMessenger
	headersSubscriber    consensus.HeadersPoolSubscriber
	delayedBroadcastData []*delayedBroadcastData
	maxDelayCacheSize    uint32
	mutDataForBroadcast  sync.RWMutex
}

// ShardChainMessengerArgs holds the arguments for creating a shardChainMessenger instance
type ShardChainMessengerArgs struct {
	CommonMessengerArgs
	HeadersSubscriber consensus.HeadersPoolSubscriber
	MaxDelayCacheSize uint32
}

// NewShardChainMessenger creates a new shardChainMessenger object
func NewShardChainMessenger(
	args ShardChainMessengerArgs,
) (*shardChainMessenger, error) {

	err := checkShardChainNilParameters(args)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:      args.Marshalizer,
		messenger:        args.Messenger,
		privateKey:       args.PrivateKey,
		shardCoordinator: args.ShardCoordinator,
		singleSigner:     args.SingleSigner,
	}

	scm := &shardChainMessenger{
		commonMessenger:      cm,
		headersSubscriber:    args.HeadersSubscriber,
		delayedBroadcastData: make([]*delayedBroadcastData, 0),
		maxDelayCacheSize:    args.MaxDelayCacheSize,
		mutDataForBroadcast:  sync.RWMutex{},
	}

	scm.headersSubscriber.RegisterHandler(scm.headerReceived)
	return scm, nil
}

func checkShardChainNilParameters(
	args ShardChainMessengerArgs,
) error {
	err := checkCommonMessengerNilParameters(args.CommonMessengerArgs)
	if err != nil {
		return err
	}
	if check.IfNil(args.HeadersSubscriber) {
		return spos.ErrNilHeadersSubscriber
	}
	if args.MaxDelayCacheSize == 0 {
		return spos.ErrInvalidCacheSize
	}

	return nil
}

// BroadcastBlock will send on in-shard headers topic and on in-shard miniblocks topic the header and block body
func (scm *shardChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if check.IfNil(blockBody) {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	b := blockBody.(*block.Body)
	msgBlockBody, err := scm.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	headerIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	selfIdentifier := scm.shardCoordinator.CommunicationIdentifier(scm.shardCoordinator.SelfId())

	go scm.messenger.Broadcast(factory.ShardBlocksTopic+headerIdentifier, msgHeader)
	go scm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *shardChainMessenger) BroadcastHeader(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	go scm.messenger.Broadcast(factory.ShardBlocksTopic+shardIdentifier, msgHeader)

	return nil
}

// SetDataForDelayBroadcast sets the miniblocks and transactions to be broadcast with delay
func (scm *shardChainMessenger) SetDataForDelayBroadcast(
	headerHash []byte,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) error {
	if len(headerHash) == 0 {
		return spos.ErrNilHeaderHash
	}
	if len(miniBlocks) == 0 {
		return nil
	}

	scm.mutDataForBroadcast.Lock()
	defer scm.mutDataForBroadcast.Unlock()

	broadcastData := &delayedBroadcastData{
		headerHash:   headerHash,
		miniblocks:   miniBlocks,
		transactions: transactions,
	}

	scm.delayedBroadcastData = append(scm.delayedBroadcastData, broadcastData)
	if len(scm.delayedBroadcastData) > int(scm.maxDelayCacheSize) {
		scm.broadcastDelayedData(scm.delayedBroadcastData[:1])
		scm.delayedBroadcastData = scm.delayedBroadcastData[1:]
	}

	return nil
}

func (scm *shardChainMessenger) headerReceived(headerHandler data.HeaderHandler, _ []byte) {
	scm.mutDataForBroadcast.RLock()
	defer scm.mutDataForBroadcast.RUnlock()

	if len(scm.delayedBroadcastData) == 0 {
		return
	}
	if headerHandler.GetShardID() != core.MetachainShardId {
		return
	}

	headerHashes, err := getShardHeaderHashesFromMetachainBlock(headerHandler, scm.shardCoordinator.SelfId())
	if err != nil {
		log.Error("notifier headerReceived", "error", err.Error())
		return
	}
	if len(headerHashes) == 0 {
		return
	}

	go scm.broadcastDataForHeaders(headerHashes)
}

func (scm *shardChainMessenger) broadcastDataForHeaders(headerHashes [][]byte) {
	time.Sleep(core.ExtraDelayForBroadcastBlockInfo)

	scm.mutDataForBroadcast.Lock()
	defer scm.mutDataForBroadcast.Unlock()

	for i := len(scm.delayedBroadcastData) - 1; i >= 0; i-- {
		for _, headerHash := range headerHashes {
			if bytes.Equal(scm.delayedBroadcastData[i].headerHash, headerHash) {
				scm.broadcastDelayedData(scm.delayedBroadcastData[:i+1])
				scm.delayedBroadcastData = scm.delayedBroadcastData[i+1:]
				return
			}
		}
	}
}

func (scm *shardChainMessenger) broadcastDelayedData(broadcastData []*delayedBroadcastData) {
	for _, bData := range broadcastData {
		err := scm.BroadcastMiniBlocks(bData.miniblocks)
		if err != nil {
			log.Error("broadcastDelayedData miniblocks", "error", err.Error())
		}

		err = scm.BroadcastTransactions(bData.transactions)
		if err != nil {
			log.Error("broadcastDelayedData transactions", "error", err.Error())
		}
	}
}

func getShardHeaderHashesFromMetachainBlock(headerHandler data.HeaderHandler, shardID uint32) ([][]byte, error) {
	metaHeader, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return nil, spos.ErrInvalidMetaHeader
	}

	shardHeaderHashes := make([][]byte, 0)
	shardsInfo := metaHeader.GetShardInfo()
	for _, shardInfo := range shardsInfo {
		if shardInfo.ShardID == shardID {
			shardHeaderHashes = append(shardHeaderHashes, shardInfo.HeaderHash)
		}
	}
	return shardHeaderHashes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardChainMessenger) IsInterfaceNil() bool {
	return scm == nil
}
