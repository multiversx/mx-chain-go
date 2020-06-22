package broadcast

import (
	"bytes"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

const prefixHeaderAlarm = "header_"
const prefixDelayDataAlarm = "delay_"
const sizeHeadersCache = 1000 // 1000 hashes in cache

// ArgsDelayedBlockBroadcaster holds the arguments to create a delayed block broadcaster
type ArgsDelayedBlockBroadcaster struct {
	InterceptorsContainer process.InterceptorsContainer
	HeadersSubscriber     consensus.HeadersPoolSubscriber
	ShardCoordinator      sharding.Coordinator
	LeaderCacheSize       uint32
	ValidatorCacheSize    uint32
}

type validatorHeaderBroadcastData struct {
	headerHash           []byte
	header               data.HeaderHandler
	metaMiniBlocksData   map[uint32][]byte
	metaTransactionsData map[string][][]byte
	order                uint32
}

type delayedBroadcastData struct {
	headerHash      []byte
	header          data.HeaderHandler
	miniBlocksData  map[uint32][]byte
	miniBlockHashes map[string]map[string]struct{}
	transactions    map[string][][]byte
	order           uint32
}

// timersScheduler exposes functionality for scheduling multiple timers
type timersScheduler interface {
	Add(callback func(alarmID string), duration time.Duration, alarmID string)
	Cancel(alarmID string)
	Close()
	IsInterfaceNil() bool
}

type headerDataForValidator struct {
	round        uint64
	prevRandSeed []byte
}

type delayedBlockBroadcaster struct {
	alarm                      timersScheduler
	interceptorsContainer      process.InterceptorsContainer
	shardCoordinator           sharding.Coordinator
	headersSubscriber          consensus.HeadersPoolSubscriber
	valHeaderBroadcastData     []*validatorHeaderBroadcastData
	valBroadcastData           []*delayedBroadcastData
	delayedBroadcastData       []*delayedBroadcastData
	maxDelayCacheSize          uint32
	maxValidatorDelayCacheSize uint32
	mutDataForBroadcast        sync.RWMutex
	broadcastMiniblocksData    func(mbData map[uint32][]byte) error
	broadcastTxsData           func(txData map[string][][]byte) error
	broadcastHeader            func(header data.HeaderHandler) error
	cacheHeaders               storage.Cacher
	mutHeadersCache            sync.RWMutex
}

// NewDelayedBlockBroadcaster create a new instance of a delayed block data broadcaster
func NewDelayedBlockBroadcaster(args *ArgsDelayedBlockBroadcaster) (*delayedBlockBroadcaster, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, spos.ErrNilShardCoordinator
	}
	if check.IfNil(args.InterceptorsContainer) {
		return nil, spos.ErrNilInterceptorsContainer
	}
	if check.IfNil(args.HeadersSubscriber) {
		return nil, spos.ErrNilHeadersSubscriber
	}

	cacheHeaders, err := lrucache.NewCache(sizeHeadersCache)
	if err != nil {
		return nil, err
	}

	dbb := &delayedBlockBroadcaster{
		alarm:                      alarm.NewAlarmScheduler(),
		shardCoordinator:           args.ShardCoordinator,
		interceptorsContainer:      args.InterceptorsContainer,
		headersSubscriber:          args.HeadersSubscriber,
		valHeaderBroadcastData:     make([]*validatorHeaderBroadcastData, 0),
		valBroadcastData:           make([]*delayedBroadcastData, 0),
		delayedBroadcastData:       make([]*delayedBroadcastData, 0),
		maxDelayCacheSize:          args.LeaderCacheSize,
		maxValidatorDelayCacheSize: args.ValidatorCacheSize,
		mutDataForBroadcast:        sync.RWMutex{},
		cacheHeaders:               cacheHeaders,
		mutHeadersCache:            sync.RWMutex{},
	}

	dbb.headersSubscriber.RegisterHandler(dbb.headerReceived)
	err = dbb.registerHeaderInterceptorCallback(dbb.interceptedHeader)
	if err != nil {
		return nil, err
	}

	err = dbb.registerMiniBlockInterceptorCallback(dbb.interceptedMiniBlockData)
	if err != nil {
		return nil, err
	}

	return dbb, nil
}

// SetLeaderData sets the data for consensus leader delayed broadcast
func (dbb *delayedBlockBroadcaster) SetLeaderData(broadcastData *delayedBroadcastData) error {
	if broadcastData == nil {
		return spos.ErrNilParameter
	}

	dbb.mutDataForBroadcast.Lock()
	defer dbb.mutDataForBroadcast.Unlock()

	log.Debug("delayedBroadcast.SetLeaderData setting leader delay data", "headerHash", broadcastData.headerHash)

	dbb.delayedBroadcastData = append(dbb.delayedBroadcastData, broadcastData)
	if len(dbb.delayedBroadcastData) > int(dbb.maxDelayCacheSize) {
		log.Debug("delayedBroadcast.SetLeaderData - leader broadcasts old data before alarm due to too much delay data",
			"headerHash", dbb.delayedBroadcastData[0].headerHash,
			"nbDelayedData", len(dbb.delayedBroadcastData),
			"maxDelayCacheSize", dbb.maxDelayCacheSize)
		dbb.broadcastDelayedData(dbb.delayedBroadcastData[:1])
		dbb.delayedBroadcastData = dbb.delayedBroadcastData[1:]
	}

	return nil
}

// SetHeaderForValidator sets the header to be broadcast by validator if leader fails to broadcast it
func (dbb *delayedBlockBroadcaster) SetHeaderForValidator(vData *validatorHeaderBroadcastData) error {
	if check.IfNil(vData.header) {
		return spos.ErrNilHeader
	}
	if len(vData.headerHash) == 0 {
		return spos.ErrNilHeaderHash
	}

	// set alarm only for validators that are aware that the block was finalized
	if len(vData.header.GetSignature()) != 0 {
		_, alreadyReceived := dbb.cacheHeaders.Get(vData.headerHash)
		if alreadyReceived {
			return nil
		}

		duration := validatorDelayPerOrder * time.Duration(vData.order)
		dbb.valHeaderBroadcastData = append(dbb.valHeaderBroadcastData, vData)
		alarmID := prefixHeaderAlarm + hex.EncodeToString(vData.headerHash)
		dbb.alarm.Add(dbb.headerAlarmExpired, duration, alarmID)
		log.Debug("delayedBroadcast.SetHeaderForValidator header alarm SET",
			"validatorConsensusOrder", vData.order,
			"headerHash", vData.headerHash,
			"alarmID-header", alarmID,
			"duration", duration)
	} else {
		log.Debug("delayedBroadcast.SetHeaderForValidator header alarm NOT SET",
			"order", vData.order,
		)
	}

	return nil
}

// SetValidatorData sets the data for consensus validator delayed broadcast
func (dbb *delayedBlockBroadcaster) SetValidatorData(broadcastData *delayedBroadcastData) error {
	if broadcastData == nil {
		return spos.ErrNilParameter
	}

	dbb.mutDataForBroadcast.Lock()
	defer dbb.mutDataForBroadcast.Unlock()

	log.Debug("delayedBroadcast.SetValidatorData setting validator delay data",
		"headerHash", broadcastData.headerHash,
		"round", broadcastData.header.GetRound(),
		"prevRandSeed", broadcastData.header.GetPrevRandSeed(),
	)

	broadcastData.miniBlockHashes = dbb.extractMiniBlockHashesCrossFromMe(broadcastData.header)
	dbb.valBroadcastData = append(dbb.valBroadcastData, broadcastData)

	if len(dbb.valBroadcastData) > int(dbb.maxValidatorDelayCacheSize) {
		alarmHeaderID := prefixHeaderAlarm + hex.EncodeToString(dbb.valBroadcastData[0].headerHash)
		alarmDelayID := prefixDelayDataAlarm + hex.EncodeToString(dbb.valBroadcastData[0].headerHash)
		dbb.alarm.Cancel(alarmHeaderID)
		dbb.alarm.Cancel(alarmDelayID)
		dbb.valBroadcastData = dbb.valBroadcastData[1:]
		log.Debug("delayedBroadcast.SetValidatorData - canceling old alarms (header and delay data) due to too much delay data,"+
			"headerHash", dbb.valBroadcastData[0].headerHash,
			"alarmID-header", alarmHeaderID,
			"alarmID-delay", alarmDelayID,
			"nbDelayData", len(dbb.valBroadcastData),
			"maxValidatorDelayCacheSize", dbb.maxValidatorDelayCacheSize)
	}

	return nil
}

// SetBroadcastHandlers sets the broadcast handlers for miniBlocks and transactions
func (dbb *delayedBlockBroadcaster) SetBroadcastHandlers(
	mbBroadcast func(mbData map[uint32][]byte) error,
	txBroadcast func(txData map[string][][]byte) error,
	headerBroadcast func(header data.HeaderHandler) error,
) error {
	if mbBroadcast == nil || txBroadcast == nil || headerBroadcast == nil {
		return spos.ErrNilParameter
	}

	dbb.mutDataForBroadcast.Lock()
	defer dbb.mutDataForBroadcast.Unlock()

	dbb.broadcastMiniblocksData = mbBroadcast
	dbb.broadcastTxsData = txBroadcast
	dbb.broadcastHeader = headerBroadcast

	return nil
}

// Close closes all the started infinite looping goroutines and subcomponents
func (dbb *delayedBlockBroadcaster) Close() {
	dbb.alarm.Close()
}

func (dbb *delayedBlockBroadcaster) headerReceived(headerHandler data.HeaderHandler, headerHash []byte) {
	dbb.mutDataForBroadcast.RLock()
	defer dbb.mutDataForBroadcast.RUnlock()

	if len(dbb.delayedBroadcastData) == 0 && len(dbb.valBroadcastData) == 0 {
		return
	}
	if headerHandler.GetShardID() != core.MetachainShardId {
		return
	}

	headerHashes, dataForValidators, err := getShardDataFromMetaChainBlock(
		headerHandler,
		dbb.shardCoordinator.SelfId(),
	)
	if err != nil {
		log.Error("delayedBroadcast.headerReceived", "error", err.Error(), "headerHash", headerHash)
		return
	}
	if len(headerHashes) == 0 {
		log.Debug("delayedBroadcast.headerReceived header received with no shardData for current shard", "headerHash", headerHash)
		return
	}

	log.Debug("delayedBroadcast.headerReceived", "nbHeaderHashes", len(headerHashes))
	for i := range headerHashes {
		log.Debug("delayedBroadcast.headerReceived", "headerHash", headerHashes[i])
	}

	go dbb.scheduleValidatorBroadcast(dataForValidators)
	go dbb.broadcastDataForHeaders(headerHashes)
}

func (dbb *delayedBlockBroadcaster) broadcastDataForHeaders(headerHashes [][]byte) {
	dbb.mutDataForBroadcast.RLock()
	if len(dbb.delayedBroadcastData) == 0 {
		dbb.mutDataForBroadcast.RUnlock()

		return
	}
	dbb.mutDataForBroadcast.RUnlock()

	time.Sleep(core.ExtraDelayForBroadcastBlockInfo)

	dbb.mutDataForBroadcast.Lock()
	defer dbb.mutDataForBroadcast.Unlock()

	for i := len(dbb.delayedBroadcastData) - 1; i >= 0; i-- {
		for _, headerHash := range headerHashes {
			if bytes.Equal(dbb.delayedBroadcastData[i].headerHash, headerHash) {
				log.Debug("delayedBroadcast.broadcastDataForHeaders - leader broadcasts block data",
					"headerHash", headerHash)
				dbb.broadcastDelayedData(dbb.delayedBroadcastData[:i+1])
				dbb.delayedBroadcastData = dbb.delayedBroadcastData[i+1:]
				return
			}
		}
	}
	log.Debug("delayedBroadcast.broadcastDataForHeaders did not find any registered data to broadcast")
}

func (dbb *delayedBlockBroadcaster) scheduleValidatorBroadcast(dataForValidators []*headerDataForValidator) {
	dbb.mutDataForBroadcast.RLock()
	defer dbb.mutDataForBroadcast.RUnlock()

	if len(dbb.valBroadcastData) == 0 {
		return
	}

	log.Debug("delayedBroadcast.scheduleValidatorBroadcast - header data for validator")
	for i := range dataForValidators {
		log.Debug("delayedBroadcast.scheduleValidatorBroadcast",
			"round", dataForValidators[i].round,
			"prevRandSeed", dataForValidators[i].prevRandSeed,
		)
	}

	log.Debug("delayedBroadcast.scheduleValidatorBroadcast - registered data for broadcast")
	for i := range dbb.valBroadcastData {
		log.Debug("delayedBroadcast.scheduleValidatorBroadcast",
			"round", dbb.valBroadcastData[i].header.GetRound(),
			"prevRandSeed", dbb.valBroadcastData[i].header.GetPrevRandSeed(),
		)
	}

	for _, headerData := range dataForValidators {
		for _, broadcastData := range dbb.valBroadcastData {
			sameRound := headerData.round == broadcastData.header.GetRound()
			samePrevRandomness := bytes.Equal(headerData.prevRandSeed, broadcastData.header.GetPrevRandSeed())
			if sameRound && samePrevRandomness {
				duration := validatorDelayPerOrder*time.Duration(broadcastData.order) + core.ExtraDelayForBroadcastBlockInfo
				alarmID := prefixDelayDataAlarm + hex.EncodeToString(broadcastData.headerHash)
				dbb.alarm.Add(dbb.alarmExpired, duration, alarmID)
				log.Debug("delayedBroadcast.scheduleValidatorBroadcast - scheduling delay data broadcast for notarized header",
					"headerHash", broadcastData.headerHash,
					"alarmID-delay", alarmID,
					"round", headerData.round,
					"prevRandSeed", headerData.prevRandSeed,
					"consensusOrder", broadcastData.order,
				)
			}
		}
	}
}

func (dbb *delayedBlockBroadcaster) alarmExpired(alarmID string) {
	headerHash, err := hex.DecodeString(strings.TrimLeft(alarmID, prefixDelayDataAlarm))
	if err != nil {
		log.Error("delayedBroadcast.alarmExpired", "error", err.Error())
		return
	}

	dbb.mutDataForBroadcast.Lock()
	defer dbb.mutDataForBroadcast.Unlock()

	for i, broadcastData := range dbb.valBroadcastData {
		if bytes.Equal(broadcastData.headerHash, headerHash) {
			log.Debug("delayedBroadcast.alarmExpired - validator broadcasts block data (with delay) instead of leader",
				"headerHash", headerHash,
				"alarmID-delay", alarmID,
			)
			dbb.broadcastDelayedData([]*delayedBroadcastData{broadcastData})
			dbb.valBroadcastData = append(dbb.valBroadcastData[:i], dbb.valBroadcastData[i+1:]...)
			return
		}
	}
}

func (dbb *delayedBlockBroadcaster) headerAlarmExpired(alarmID string) {
	headerHash, err := hex.DecodeString(strings.TrimPrefix(alarmID, prefixHeaderAlarm))
	if err != nil {
		log.Error("delayedBroadcast.headerAlarmExpired", "error", err.Error())
		return
	}

	dbb.mutDataForBroadcast.Lock()
	defer dbb.mutDataForBroadcast.Unlock()

	var vHeader *validatorHeaderBroadcastData
	for i, broadcastData := range dbb.valHeaderBroadcastData {
		if bytes.Equal(broadcastData.headerHash, headerHash) {
			vHeader = broadcastData
			dbb.valHeaderBroadcastData = append(dbb.valHeaderBroadcastData[:i], dbb.valHeaderBroadcastData[i+1:]...)
			break
		}
	}

	if vHeader == nil {
		log.Warn("delayedBroadcast.headerAlarmExpired", "error", "alarm data is nil")
		return
	}

	log.Debug("delayedBroadcast.headerAlarmExpired - validator broadcasting header",
		"headerHash", headerHash,
		"alarmID-header", alarmID,
	)
	// broadcast header
	err = dbb.broadcastHeader(vHeader.header)
	if err != nil {
		log.Warn("delayedBroadcast.headerAlarmExpired", "error", err.Error())
	}

	// if metaChain broadcast meta data with extra delay
	if dbb.shardCoordinator.SelfId() == core.MetachainShardId {
		log.Debug("delayedBroadcast.headerAlarmExpired - validator broadcasting meta miniblocks and transactions",
			"headerHash", headerHash,
		)
		dbb.broadcastBlockData(vHeader.metaMiniBlocksData, vHeader.metaTransactionsData, core.ExtraDelayForBroadcastBlockInfo)
	}
}

func (dbb *delayedBlockBroadcaster) broadcastDelayedData(broadcastData []*delayedBroadcastData) {
	for _, bData := range broadcastData {
		dbb.broadcastBlockData(bData.miniBlocksData, bData.transactions, 0)
	}
}

func (dbb *delayedBlockBroadcaster) broadcastBlockData(
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	delay time.Duration,
) {
	time.Sleep(delay)

	err := dbb.broadcastMiniblocksData(miniBlocks)
	if err != nil {
		log.Error("broadcastBlockData miniblocks", "error", err.Error())
	}

	err = dbb.broadcastTxsData(transactions)
	if err != nil {
		log.Error("broadcastBlockData transactions", "error", err.Error())
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

func (dbb *delayedBlockBroadcaster) registerHeaderInterceptorCallback(
	cb func(topic string, hash []byte, data interface{}),
) error {
	selfShardID := dbb.shardCoordinator.SelfId()

	if selfShardID == core.MetachainShardId {
		return dbb.registerInterceptorsCallbackForMetaHeader(cb)
	}

	identifierShardHeader := factory.ShardBlocksTopic + dbb.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	interceptor, err := dbb.interceptorsContainer.Get(identifierShardHeader)
	if err != nil {
		return err
	}

	interceptor.RegisterHandler(cb)
	return nil
}

func (dbb *delayedBlockBroadcaster) registerMiniBlockInterceptorCallback(
	cb func(topic string, hash []byte, data interface{}),
) error {
	if dbb.shardCoordinator.SelfId() == core.MetachainShardId {
		return dbb.registerInterceptorsCallbackForMetaMiniblocks(cb)
	}

	return dbb.registerInterceptorsCallbackForShard(factory.MiniBlocksTopic, cb)
}

func (dbb *delayedBlockBroadcaster) registerInterceptorsCallbackForMetaHeader(
	cb func(topic string, hash []byte, data interface{}),
) error {
	identifier := factory.MetachainBlocksTopic
	interceptor, err := dbb.interceptorsContainer.Get(identifier)
	if err != nil {
		return err
	}

	interceptor.RegisterHandler(cb)

	return nil
}

func (dbb *delayedBlockBroadcaster) registerInterceptorsCallbackForMetaMiniblocks(
	cb func(topic string, hash []byte, data interface{}),
) error {
	identifier := factory.MiniBlocksTopic + dbb.shardCoordinator.CommunicationIdentifier(core.AllShardId)
	interceptor, err := dbb.interceptorsContainer.Get(identifier)
	if err != nil {
		return err
	}

	interceptor.RegisterHandler(cb)

	return nil
}

func (dbb *delayedBlockBroadcaster) registerInterceptorsCallbackForShard(
	rootTopic string,
	cb func(topic string, hash []byte, data interface{}),
) error {
	for idx := uint32(0); idx < dbb.shardCoordinator.NumberOfShards(); idx++ {
		// interested only in cross shard data
		if idx == dbb.shardCoordinator.SelfId() {
			continue
		}
		identifierMiniBlocks := rootTopic + dbb.shardCoordinator.CommunicationIdentifier(idx)
		interceptor, err := dbb.interceptorsContainer.Get(identifierMiniBlocks)
		if err != nil {
			return err
		}

		interceptor.RegisterHandler(cb)
	}

	return nil
}

func (dbb *delayedBlockBroadcaster) interceptedHeader(_ string, headerHash []byte, header interface{}) {
	headerHandler, ok := header.(data.HeaderHandler)
	if !ok {
		log.Warn("delayedBroadcast.interceptedHeader",
			"error", "not a header")
		return
	}

	dbb.mutHeadersCache.Lock()
	dbb.cacheHeaders.Put(headerHash, struct{}{}, 0)
	dbb.mutHeadersCache.Unlock()

	dbb.mutDataForBroadcast.RLock()
	defer dbb.mutDataForBroadcast.RUnlock()

	log.Debug("delayedBroadcast.interceptedHeader",
		"headerHash", headerHash,
		"round", headerHandler.GetRound(),
		"prevRandSeed", headerHandler.GetPrevRandSeed(),
	)

	for _, broadcastData := range dbb.valHeaderBroadcastData {
		samePrevRandSeed := bytes.Equal(broadcastData.header.GetPrevRandSeed(), headerHandler.GetPrevRandSeed())
		sameRound := broadcastData.header.GetRound() == headerHandler.GetRound()
		sameHeader := samePrevRandSeed && sameRound

		if sameHeader {
			// leader has broadcast the header so we can cancel the header alarm
			alarmID := prefixHeaderAlarm + hex.EncodeToString(headerHash)
			dbb.alarm.Cancel(alarmID)
			log.Debug("delayedBroadcast.interceptedHeader canceling alarm for broadcasting header",
				"headerHash", headerHash,
				"alarmID-header", alarmID,
			)
			break
		}
	}
}

func (dbb *delayedBlockBroadcaster) interceptedMiniBlockData(topic string, hash []byte, _ interface{}) {
	dbb.mutDataForBroadcast.Lock()
	defer dbb.mutDataForBroadcast.Unlock()

	remainingValBroadcastData := make([]*delayedBroadcastData, 0)
	for i, broadcastData := range dbb.valBroadcastData {
		mbHashesMap := broadcastData.miniBlockHashes
		if len(mbHashesMap) > 0 && len(mbHashesMap[topic]) > 0 {
			delete(broadcastData.miniBlockHashes[topic], string(hash))
			if len(mbHashesMap[topic]) == 0 {
				delete(mbHashesMap, topic)
			}
		}

		if len(mbHashesMap) == 0 {
			alarmID := prefixDelayDataAlarm + hex.EncodeToString(broadcastData.headerHash)
			dbb.alarm.Cancel(alarmID)
			log.Debug("delayedBroadcast.interceptedMiniBlockData leader has broadcast block data, validator cancelling alarm",
				"headerHash", broadcastData.headerHash,
				"alarmID-delay", alarmID,
			)
		} else {
			remainingValBroadcastData = append(remainingValBroadcastData, dbb.valBroadcastData[i])
		}
	}
	dbb.valBroadcastData = remainingValBroadcastData
}

func (dbb *delayedBlockBroadcaster) extractMiniBlockHashesCrossFromMe(header data.HeaderHandler) map[string]map[string]struct{} {
	shardID := dbb.shardCoordinator.SelfId()
	mbHashesForShards := make(map[string]map[string]struct{})
	for i := uint32(0); i < dbb.shardCoordinator.NumberOfShards(); i++ {
		if i == shardID {
			continue
		}
		topic := factory.MiniBlocksTopic + dbb.shardCoordinator.CommunicationIdentifier(i)
		mbs := dbb.extractMbsFromMeTo(header, i)
		if len(mbs) == 0 {
			continue
		}
		mbHashesForShards[topic] = mbs
	}

	if shardID != core.MetachainShardId {
		topic := factory.MiniBlocksTopic + dbb.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
		mbs := dbb.extractMbsFromMeTo(header, core.MetachainShardId)
		if len(mbs) > 0 {
			mbHashesForShards[topic] = mbs
		}
	}

	return mbHashesForShards
}

func (dbb *delayedBlockBroadcaster) extractMbsFromMeTo(header data.HeaderHandler, toShardID uint32) map[string]struct{} {
	mbHashesForShard := make(map[string]struct{})
	miniBlockHeaders := header.GetMiniBlockHeadersWithDst(toShardID)
	for key := range miniBlockHeaders {
		mbHashesForShard[key] = struct{}{}
	}

	return mbHashesForShard
}
