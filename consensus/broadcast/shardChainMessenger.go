package broadcast

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

const validatorDelayPerOrder = time.Second

var _ consensus.BroadcastMessenger = (*shardChainMessenger)(nil)

type delayedBroadcastData struct {
	headerHash      []byte
	prevRandSeed    []byte
	round           uint64
	miniblocks      map[uint32][]byte
	miniBlockHashes map[uint32]map[string]struct{}
	transactions    map[string][][]byte
	order           uint32
}

type shardChainMessenger struct {
	*commonMessenger
	headersSubscriber          consensus.HeadersPoolSubscriber
	valBroadcastData           []*delayedBroadcastData
	delayedBroadcastData       []*delayedBroadcastData
	maxDelayCacheSize          uint32
	maxValidatorDelayCacheSize uint32
	mutDataForBroadcast        sync.RWMutex
}

// ShardChainMessengerArgs holds the arguments for creating a shardChainMessenger instance
type ShardChainMessengerArgs struct {
	CommonMessengerArgs
	HeadersSubscriber          consensus.HeadersPoolSubscriber
	MaxDelayCacheSize          uint32
	MaxValidatorDelayCacheSize uint32
}

type headerDataForValidator struct {
	round        uint64
	prevRandSeed []byte
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
		marshalizer:           args.Marshalizer,
		messenger:             args.Messenger,
		privateKey:            args.PrivateKey,
		shardCoordinator:      args.ShardCoordinator,
		singleSigner:          args.SingleSigner,
		alarm:                 alarm.NewAlarmScheduler(),
		interceptorsContainer: args.InterceptorsContainer,
	}

	scm := &shardChainMessenger{
		commonMessenger:            cm,
		headersSubscriber:          args.HeadersSubscriber,
		valBroadcastData:           make([]*delayedBroadcastData, 0),
		delayedBroadcastData:       make([]*delayedBroadcastData, 0),
		maxDelayCacheSize:          args.MaxDelayCacheSize,
		maxValidatorDelayCacheSize: args.MaxValidatorDelayCacheSize,
		mutDataForBroadcast:        sync.RWMutex{},
	}

	scm.headersSubscriber.RegisterHandler(scm.headerReceived)
	err = scm.registerInterceptorCallback(scm.interceptedMiniBlockData)
	if err != nil {
		return nil, err
	}

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

// SetLeaderDelayBroadcast sets the miniBlocks and transactions to be broadcast with delay
func (scm *shardChainMessenger) SetLeaderDelayBroadcast(
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

// SetValidatorDelayBroadcast sets the miniBlocks and transactions to be broadcast with delay
// The broadcast will only be done in case the consensus leader fails its broadcast
// and all other validators with a lower order fail their broadcast as well.
func (scm *shardChainMessenger) SetValidatorDelayBroadcast(headerHash []byte, prevRandSeed []byte, round uint64, miniBlocks map[uint32][]byte, miniBlockHashes map[uint32]map[string]struct{}, transactions map[string][][]byte, order uint32) error {
	if len(headerHash) == 0 {
		return spos.ErrNilHeaderHash
	}
	if len(prevRandSeed) == 0 {
		return spos.ErrNilPrevRandSeed
	}
	if len(miniBlocks) == 0 && len(miniBlockHashes) == 0 {
		return nil
	}
	if len(miniBlocks) == 0 || len(miniBlockHashes) == 0 {
		return spos.ErrInvalidDataToBroadcast
	}

	scm.mutDataForBroadcast.Lock()
	defer scm.mutDataForBroadcast.Unlock()

	broadcastData := &delayedBroadcastData{
		headerHash:      headerHash,
		prevRandSeed:    prevRandSeed,
		round:           round,
		miniblocks:      miniBlocks,
		miniBlockHashes: miniBlockHashes,
		transactions:    transactions,
		order:           order,
	}

	scm.valBroadcastData = append(scm.valBroadcastData, broadcastData)
	if len(scm.valBroadcastData) > int(scm.maxValidatorDelayCacheSize) {
		scm.valBroadcastData = scm.valBroadcastData[1:]
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

	headerHashes, dataForValidators, err := getShardDataFromMetaChainBlock(headerHandler, scm.shardCoordinator.SelfId())
	if err != nil {
		log.Error("notifier headerReceived", "error", err.Error())
		return
	}
	if len(headerHashes) == 0 {
		return
	}

	go scm.scheduleValidatorBroadcast(dataForValidators)

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

func (scm *shardChainMessenger) scheduleValidatorBroadcast(dataForValidators []*headerDataForValidator) {
	for _, headerData := range dataForValidators {
		for _, broadcastData := range scm.valBroadcastData {
			sameRound := headerData.round == broadcastData.round
			samePrevRandomness := bytes.Equal(headerData.prevRandSeed, broadcastData.prevRandSeed)
			if sameRound && samePrevRandomness {
				duration := validatorDelayPerOrder * time.Duration(broadcastData.order)
				scm.alarm.Add(scm.alarmExpired, duration, string(broadcastData.headerHash))
			}
		}
	}
}

func (scm *shardChainMessenger) alarmExpired(headerHash string) {
	scm.mutDataForBroadcast.Lock()
	defer scm.mutDataForBroadcast.Unlock()
	for i, broadcastData := range scm.valBroadcastData {
		if bytes.Equal(broadcastData.headerHash, []byte(headerHash)) {
			scm.broadcastDelayedData([]*delayedBroadcastData{broadcastData})
			scm.valBroadcastData = append(scm.valBroadcastData[:i], scm.valBroadcastData[i+1:]...)
			return
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

func getShardDataFromMetaChainBlock(
	headerHandler data.HeaderHandler,
	shardID uint32,
) ([][]byte, []*headerDataForValidator, error) {
	metaHeader, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return nil, nil, spos.ErrInvalidMetaHeader
	}

	dataForValidators := make([]*headerDataForValidator, 0)
	shardHeaderHashes := make([][]byte, 0)
	shardsInfo := metaHeader.GetShardInfo()
	for _, shardInfo := range shardsInfo {
		if shardInfo.ShardID == shardID {
			shardHeaderHashes = append(shardHeaderHashes, shardInfo.HeaderHash)
			headerData := &headerDataForValidator{
				round:        shardInfo.GetRound(),
				prevRandSeed: shardInfo.GetPrevRandSeed(),
			}
			dataForValidators = append(dataForValidators, headerData)
		}
	}
	return shardHeaderHashes, dataForValidators, nil
}

func (scm *shardChainMessenger) registerInterceptorCallback(cb func(toShard uint32, data []byte)) error {
	for idx := uint32(0); idx < scm.shardCoordinator.NumberOfShards(); idx++ {
		// interested only in cross shard data
		if idx == scm.shardCoordinator.SelfId() {
			continue
		}
		identifierMiniBlocks := factory.MiniBlocksTopic + scm.shardCoordinator.CommunicationIdentifier(idx)
		interceptor, err := scm.interceptorsContainer.Get(identifierMiniBlocks)
		if err != nil {
			return err
		}

		interceptor.RegisterHandler(cb)
	}

	return nil
}

func (scm *shardChainMessenger) interceptedMiniBlockData(toShardTopic uint32, data []byte) {
	scm.mutDataForBroadcast.Lock()
	defer scm.mutDataForBroadcast.Unlock()

	for i, broadcastData := range scm.valBroadcastData {
		mbHashesMap := broadcastData.miniBlockHashes
		if len(mbHashesMap) > 0 && len(mbHashesMap[toShardTopic]) > 0 {
			delete(broadcastData.miniBlockHashes[toShardTopic], string(data))
			if len(mbHashesMap[toShardTopic]) == 0 {
				delete(mbHashesMap, toShardTopic)
			}
		}

		if len(mbHashesMap) == 0 {
			scm.alarm.Cancel(string(broadcastData.headerHash))
			scm.valBroadcastData = append(scm.valBroadcastData[:i], scm.valBroadcastData[i+1:]...)
			return
		}
	}
}

// Close closes all its started infinite looping goroutines
func (scm *shardChainMessenger) Close() {
	scm.alarm.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardChainMessenger) IsInterfaceNil() bool {
	return scm == nil
}
