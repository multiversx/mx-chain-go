package interceptors

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/pkg/errors"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/debug/handler"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors/disabled"
)

var log = logger.GetOrCreate("process/interceptors")

// ArgMultiDataInterceptor is the argument for the multi-data interceptor
type ArgMultiDataInterceptor struct {
	Topic                   string
	Marshalizer             marshal.Marshalizer
	Hasher                  hashing.Hasher
	DataFactory             process.InterceptedDataFactory
	Processor               process.InterceptorProcessor
	Throttler               process.InterceptorThrottler
	AntifloodHandler        process.P2PAntifloodHandler
	WhiteListRequest        process.WhiteListHandler
	PreferredPeersHolder    process.PreferredPeersHolderHandler
	CurrentPeerId           core.PeerID
	InterceptedDataVerifier process.InterceptedDataVerifier
}

// MultiDataInterceptor is used for intercepting packed multi data
type MultiDataInterceptor struct {
	*baseDataInterceptor
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	factory            process.InterceptedDataFactory
	whiteListRequest   process.WhiteListHandler
	mutChunksProcessor sync.RWMutex
	chunksProcessor    process.InterceptedChunksProcessor
}

// NewMultiDataInterceptor hooks a new interceptor for packed multi data
func NewMultiDataInterceptor(arg ArgMultiDataInterceptor) (*MultiDataInterceptor, error) {
	if len(arg.Topic) == 0 {
		return nil, process.ErrEmptyTopic
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(arg.DataFactory) {
		return nil, process.ErrNilInterceptedDataFactory
	}
	if check.IfNil(arg.Processor) {
		return nil, process.ErrNilInterceptedDataProcessor
	}
	if check.IfNil(arg.Throttler) {
		return nil, process.ErrNilInterceptorThrottler
	}
	if check.IfNil(arg.AntifloodHandler) {
		return nil, process.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.WhiteListRequest) {
		return nil, process.ErrNilWhiteListHandler
	}
	if check.IfNil(arg.PreferredPeersHolder) {
		return nil, process.ErrNilPreferredPeersHolder
	}
	if check.IfNil(arg.InterceptedDataVerifier) {
		return nil, process.ErrNilInterceptedDataVerifier
	}
	if len(arg.CurrentPeerId) == 0 {
		return nil, process.ErrEmptyPeerID
	}

	multiDataIntercept := &MultiDataInterceptor{
		baseDataInterceptor: &baseDataInterceptor{
			throttler:               arg.Throttler,
			antifloodHandler:        arg.AntifloodHandler,
			topic:                   arg.Topic,
			currentPeerId:           arg.CurrentPeerId,
			processor:               arg.Processor,
			preferredPeersHolder:    arg.PreferredPeersHolder,
			debugHandler:            handler.NewDisabledInterceptorDebugHandler(),
			interceptedDataVerifier: arg.InterceptedDataVerifier,
		},
		marshalizer:      arg.Marshalizer,
		hasher:           arg.Hasher,
		factory:          arg.DataFactory,
		whiteListRequest: arg.WhiteListRequest,
		chunksProcessor:  disabled.NewDisabledInterceptedChunksProcessor(),
	}

	return multiDataIntercept, nil
}

// ProcessReceivedMessage is the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (mdi *MultiDataInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
	err := mdi.preProcessMesage(message, fromConnectedPeer)
	if err != nil {
		return nil, err
	}

	b := batch.Batch{}
	err = mdi.marshalizer.Unmarshal(&b, message.Data())
	if err != nil {
		mdi.throttler.EndProcessing()

		// this situation is so severe that we need to black list de peers
		reason := "unmarshalable data got on topic " + mdi.topic
		mdi.antifloodHandler.BlacklistPeer(message.Peer(), reason, common.InvalidMessageBlacklistDuration)
		mdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, common.InvalidMessageBlacklistDuration)

		return nil, err
	}
	multiDataBuff := b.Data
	lenMultiData := len(multiDataBuff)
	if lenMultiData == 0 {
		mdi.throttler.EndProcessing()
		return nil, process.ErrNoDataInMessage
	}

	err = mdi.antifloodHandler.CanProcessMessagesOnTopic(
		fromConnectedPeer,
		mdi.topic,
		uint32(lenMultiData),
		uint64(len(message.Data())),
		message.SeqNo(),
	)
	if err != nil {
		mdi.throttler.EndProcessing()
		return nil, err
	}

	mdi.mutChunksProcessor.RLock()
	checkChunksRes, err := mdi.chunksProcessor.CheckBatch(&b, mdi.whiteListRequest)
	mdi.mutChunksProcessor.RUnlock()
	if err != nil {
		mdi.throttler.EndProcessing()
		return nil, err
	}

	isIncompleteChunk := checkChunksRes.IsChunk && !checkChunksRes.HaveAllChunks
	if isIncompleteChunk {
		mdi.throttler.EndProcessing()
		return nil, nil
	}
	isCompleteChunk := checkChunksRes.IsChunk && checkChunksRes.HaveAllChunks
	if isCompleteChunk {
		multiDataBuff = [][]byte{checkChunksRes.CompleteBuffer}
	}

	listInterceptedData := make([]process.InterceptedData, len(multiDataBuff))
	errOriginator := mdi.antifloodHandler.IsOriginatorEligibleForTopic(message.Peer(), mdi.topic)

	for index, dataBuff := range multiDataBuff {
		var interceptedData process.InterceptedData
		interceptedData, err = mdi.interceptedData(dataBuff, message.Peer(), fromConnectedPeer)
		listInterceptedData[index] = interceptedData

		if err != nil {
			mdi.throttler.EndProcessing()
			return nil, err
		}

		isWhiteListed := mdi.whiteListRequest.IsWhiteListed(interceptedData)
		if !isWhiteListed && errOriginator != nil {
			mdi.throttler.EndProcessing()
			log.Trace("got message from peer on topic only for validators", "originator",
				p2p.PeerIdToShortString(message.Peer()),
				"topic", mdi.topic,
				"err", errOriginator)
			return nil, errOriginator
		}

		isForCurrentShard := interceptedData.IsForCurrentShard()
		shouldProcess := isForCurrentShard || isWhiteListed
		if !shouldProcess {
			log.Trace("intercepted data should not be processed",
				"pid", p2p.MessageOriginatorPid(message),
				"seq no", p2p.MessageOriginatorSeq(message),
				"topic", message.Topic(),
				"hash", interceptedData.Hash(),
				"is for this shard", isForCurrentShard,
				"is white listed", isWhiteListed,
			)
			mdi.throttler.EndProcessing()
			return nil, process.ErrInterceptedDataNotForCurrentShard
		}
	}

	go func() {
		for _, interceptedData := range listInterceptedData {
			mdi.processInterceptedData(interceptedData, message)
		}
		mdi.throttler.EndProcessing()
	}()

	messageID := mdi.createInterceptedMultiDataMsgID(listInterceptedData)

	return messageID, nil
}

func (mdi *MultiDataInterceptor) createInterceptedMultiDataMsgID(interceptedMultiData []process.InterceptedData) []byte {
	if len(interceptedMultiData) == 0 {
		return nil
	}
	if len(interceptedMultiData) == 1 {
		return interceptedMultiData[0].Hash()
	}

	lenOneID := len(interceptedMultiData[0].Hash())
	data := make([]byte, 0, lenOneID*len(interceptedMultiData))
	for _, id := range interceptedMultiData {
		data = append(data, id.Hash()...)
	}
	if len(data) == 0 {
		return []byte{}
	}

	return mdi.hasher.Compute(string(data))
}

func (mdi *MultiDataInterceptor) interceptedData(dataBuff []byte, originator core.PeerID, fromConnectedPeer core.PeerID) (process.InterceptedData, error) {
	interceptedData, err := mdi.factory.Create(dataBuff, originator)
	if err != nil {
		// this situation is so severe that we need to black list de peers
		reason := "can not create object from received bytes, topic " + mdi.topic + ", error " + err.Error()
		mdi.antifloodHandler.BlacklistPeer(originator, reason, common.InvalidMessageBlacklistDuration)
		mdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, common.InvalidMessageBlacklistDuration)

		return nil, err
	}

	mdi.receivedDebugInterceptedData(interceptedData)

	err = mdi.interceptedDataVerifier.Verify(interceptedData)
	if err != nil {
		mdi.processDebugInterceptedData(interceptedData, err)

		isWrongVersion := errors.Is(err, process.ErrInvalidTransactionVersion) || errors.Is(err, process.ErrInvalidChainID)
		if isWrongVersion {
			// this situation is so severe that we need to black list de peers
			reason := "wrong version of received intercepted data, topic " + mdi.topic + ", error " + err.Error()
			mdi.antifloodHandler.BlacklistPeer(originator, reason, common.InvalidMessageBlacklistDuration)
			mdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, common.InvalidMessageBlacklistDuration)
		}

		return nil, err
	}

	return interceptedData, nil
}

// RegisterHandler registers a callback function to be notified on received data
func (mdi *MultiDataInterceptor) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {
	mdi.processor.RegisterHandler(handler)
}

// SetChunkProcessor sets the intercepted chunks processor
func (mdi *MultiDataInterceptor) SetChunkProcessor(processor process.InterceptedChunksProcessor) error {
	if check.IfNil(processor) {
		return process.ErrNilChunksProcessor
	}

	mdi.mutChunksProcessor.Lock()
	mdi.chunksProcessor = processor
	mdi.mutChunksProcessor.Unlock()

	return nil
}

// Close will call the chunk processor's close method
func (mdi *MultiDataInterceptor) Close() error {
	mdi.mutChunksProcessor.RLock()
	defer mdi.mutChunksProcessor.RUnlock()

	return mdi.chunksProcessor.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdi *MultiDataInterceptor) IsInterfaceNil() bool {
	return mdi == nil
}
