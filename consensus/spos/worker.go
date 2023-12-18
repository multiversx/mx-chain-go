package spos

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	errorsErd "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

var _ closing.Closer = (*Worker)(nil)

// sleepTime defines the time in milliseconds between each iteration made in checkChannels method
const sleepTime = 5 * time.Millisecond

// Worker defines the data needed by spos to communicate between nodes which are in the validators group
type Worker struct {
	consensusService        ConsensusService
	blockChain              data.ChainHandler
	blockProcessor          process.BlockProcessor
	scheduledProcessor      consensus.ScheduledProcessor
	bootstrapper            process.Bootstrapper
	broadcastMessenger      consensus.BroadcastMessenger
	consensusState          *ConsensusState
	forkDetector            process.ForkDetector
	marshalizer             marshal.Marshalizer
	hasher                  hashing.Hasher
	roundHandler            consensus.RoundHandler
	shardCoordinator        sharding.Coordinator
	peerSignatureHandler    crypto.PeerSignatureHandler
	syncTimer               ntp.SyncTimer
	headerSigVerifier       HeaderSigVerifier
	headerIntegrityVerifier process.HeaderIntegrityVerifier
	appStatusHandler        core.AppStatusHandler
	enableEpochsHandler     common.EnableEpochsHandler

	networkShardingCollector consensus.NetworkShardingCollector

	receivedMessages      map[consensus.MessageType][]*consensus.Message
	receivedMessagesCalls map[consensus.MessageType]func(ctx context.Context, msg *consensus.Message) bool

	executeMessageChannel        chan *consensus.Message
	consensusStateChangedChannel chan bool

	mutReceivedMessages      sync.RWMutex
	mutReceivedMessagesCalls sync.RWMutex

	mapDisplayHashConsensusMessage map[string][]*consensus.Message
	mutDisplayHashConsensusMessage sync.RWMutex

	receivedHeadersHandlers   []func(headerHandler data.HeaderHandler)
	mutReceivedHeadersHandler sync.RWMutex

	antifloodHandler consensus.P2PAntifloodHandler
	poolAdder        PoolAdder

	cancelFunc                func()
	consensusMessageValidator *consensusMessageValidator
	nodeRedundancyHandler     consensus.NodeRedundancyHandler
	peerBlacklistHandler      consensus.PeerBlacklistHandler
	closer                    core.SafeCloser

	mutEquivalentMessages      sync.RWMutex
	equivalentMessages         map[string]*consensus.EquivalentMessageInfo
	equivalentMessagesDebugger EquivalentMessagesDebugger
}

// WorkerArgs holds the consensus worker arguments
type WorkerArgs struct {
	ConsensusService           ConsensusService
	BlockChain                 data.ChainHandler
	BlockProcessor             process.BlockProcessor
	ScheduledProcessor         consensus.ScheduledProcessor
	Bootstrapper               process.Bootstrapper
	BroadcastMessenger         consensus.BroadcastMessenger
	ConsensusState             *ConsensusState
	ForkDetector               process.ForkDetector
	Marshalizer                marshal.Marshalizer
	Hasher                     hashing.Hasher
	RoundHandler               consensus.RoundHandler
	ShardCoordinator           sharding.Coordinator
	PeerSignatureHandler       crypto.PeerSignatureHandler
	SyncTimer                  ntp.SyncTimer
	HeaderSigVerifier          HeaderSigVerifier
	HeaderIntegrityVerifier    process.HeaderIntegrityVerifier
	ChainID                    []byte
	NetworkShardingCollector   consensus.NetworkShardingCollector
	AntifloodHandler           consensus.P2PAntifloodHandler
	PoolAdder                  PoolAdder
	SignatureSize              int
	PublicKeySize              int
	AppStatusHandler           core.AppStatusHandler
	NodeRedundancyHandler      consensus.NodeRedundancyHandler
	PeerBlacklistHandler       consensus.PeerBlacklistHandler
	EquivalentMessagesDebugger EquivalentMessagesDebugger
	EnableEpochsHandler        common.EnableEpochsHandler
}

// NewWorker creates a new Worker object
func NewWorker(args *WorkerArgs) (*Worker, error) {
	err := checkNewWorkerParams(args)
	if err != nil {
		return nil, err
	}

	argsConsensusMessageValidator := ArgsConsensusMessageValidator{
		ConsensusState:       args.ConsensusState,
		ConsensusService:     args.ConsensusService,
		PeerSignatureHandler: args.PeerSignatureHandler,
		SignatureSize:        args.SignatureSize,
		PublicKeySize:        args.PublicKeySize,
		HeaderHashSize:       args.Hasher.Size(),
		ChainID:              args.ChainID,
	}

	consensusMessageValidatorObj, err := NewConsensusMessageValidator(argsConsensusMessageValidator)
	if err != nil {
		return nil, err
	}

	wrk := Worker{
		consensusService:           args.ConsensusService,
		blockChain:                 args.BlockChain,
		blockProcessor:             args.BlockProcessor,
		scheduledProcessor:         args.ScheduledProcessor,
		bootstrapper:               args.Bootstrapper,
		broadcastMessenger:         args.BroadcastMessenger,
		consensusState:             args.ConsensusState,
		forkDetector:               args.ForkDetector,
		marshalizer:                args.Marshalizer,
		hasher:                     args.Hasher,
		roundHandler:               args.RoundHandler,
		shardCoordinator:           args.ShardCoordinator,
		peerSignatureHandler:       args.PeerSignatureHandler,
		syncTimer:                  args.SyncTimer,
		headerSigVerifier:          args.HeaderSigVerifier,
		headerIntegrityVerifier:    args.HeaderIntegrityVerifier,
		appStatusHandler:           args.AppStatusHandler,
		networkShardingCollector:   args.NetworkShardingCollector,
		antifloodHandler:           args.AntifloodHandler,
		poolAdder:                  args.PoolAdder,
		nodeRedundancyHandler:      args.NodeRedundancyHandler,
		peerBlacklistHandler:       args.PeerBlacklistHandler,
		closer:                     closing.NewSafeChanCloser(),
		equivalentMessages:         make(map[string]*consensus.EquivalentMessageInfo),
		equivalentMessagesDebugger: args.EquivalentMessagesDebugger,
		enableEpochsHandler:        args.EnableEpochsHandler,
	}

	wrk.consensusMessageValidator = consensusMessageValidatorObj
	wrk.executeMessageChannel = make(chan *consensus.Message)
	wrk.receivedMessagesCalls = make(map[consensus.MessageType]func(context.Context, *consensus.Message) bool)
	wrk.receivedHeadersHandlers = make([]func(data.HeaderHandler), 0)
	wrk.consensusStateChangedChannel = make(chan bool, 1)
	wrk.bootstrapper.AddSyncStateListener(wrk.receivedSyncState)
	wrk.initReceivedMessages()

	// set the limit for the antiflood handler
	topic := GetConsensusTopicID(args.ShardCoordinator)
	maxMessagesInARoundPerPeer := wrk.consensusService.GetMaxMessagesInARoundPerPeer()
	wrk.antifloodHandler.SetMaxMessagesForTopic(topic, maxMessagesInARoundPerPeer)

	wrk.mapDisplayHashConsensusMessage = make(map[string][]*consensus.Message)

	return &wrk, nil
}

// StartWorking actually starts the consensus working mechanism
func (wrk *Worker) StartWorking() {
	var ctx context.Context
	ctx, wrk.cancelFunc = context.WithCancel(context.Background())
	go wrk.checkChannels(ctx)
}

func checkNewWorkerParams(args *WorkerArgs) error {
	if args == nil {
		return ErrNilWorkerArgs
	}
	if check.IfNil(args.ConsensusService) {
		return ErrNilConsensusService
	}
	if check.IfNil(args.BlockChain) {
		return ErrNilBlockChain
	}
	if check.IfNil(args.BlockProcessor) {
		return ErrNilBlockProcessor
	}
	if check.IfNil(args.ScheduledProcessor) {
		return ErrNilScheduledProcessor
	}
	if check.IfNil(args.Bootstrapper) {
		return ErrNilBootstrapper
	}
	if check.IfNil(args.BroadcastMessenger) {
		return ErrNilBroadcastMessenger
	}
	if args.ConsensusState == nil {
		return ErrNilConsensusState
	}
	if check.IfNil(args.ForkDetector) {
		return ErrNilForkDetector
	}
	if check.IfNil(args.Marshalizer) {
		return ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return ErrNilHasher
	}
	if check.IfNil(args.RoundHandler) {
		return ErrNilRoundHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return ErrNilShardCoordinator
	}
	if check.IfNil(args.PeerSignatureHandler) {
		return ErrNilPeerSignatureHandler
	}
	if check.IfNil(args.SyncTimer) {
		return ErrNilSyncTimer
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return ErrNilHeaderSigVerifier
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return ErrNilHeaderIntegrityVerifier
	}
	if len(args.ChainID) == 0 {
		return ErrInvalidChainID
	}
	if check.IfNil(args.NetworkShardingCollector) {
		return ErrNilNetworkShardingCollector
	}
	if check.IfNil(args.AntifloodHandler) {
		return ErrNilAntifloodHandler
	}
	if check.IfNil(args.PoolAdder) {
		return ErrNilPoolAdder
	}
	if check.IfNil(args.AppStatusHandler) {
		return ErrNilAppStatusHandler
	}
	if check.IfNil(args.NodeRedundancyHandler) {
		return ErrNilNodeRedundancyHandler
	}
	if check.IfNil(args.PeerBlacklistHandler) {
		return ErrNilPeerBlacklistHandler
	}
	if check.IfNil(args.EquivalentMessagesDebugger) {
		return ErrNilEquivalentMessagesDebugger
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return ErrNilEnableEpochsHandler
	}

	return nil
}

func (wrk *Worker) receivedSyncState(isNodeSynchronized bool) {
	if isNodeSynchronized {
		select {
		case wrk.consensusStateChangedChannel <- true:
		default:
		}
	}
}

// ReceivedHeader process the received header, calling each received header handler registered in worker instance
func (wrk *Worker) ReceivedHeader(headerHandler data.HeaderHandler, _ []byte) {
	isHeaderForOtherShard := headerHandler.GetShardID() != wrk.shardCoordinator.SelfId()
	isHeaderForOtherRound := int64(headerHandler.GetRound()) != wrk.roundHandler.Index()
	headerCanNotBeProcessed := isHeaderForOtherShard || isHeaderForOtherRound
	if headerCanNotBeProcessed {
		return
	}

	wrk.mutReceivedHeadersHandler.RLock()
	for _, handler := range wrk.receivedHeadersHandlers {
		handler(headerHandler)
	}
	wrk.mutReceivedHeadersHandler.RUnlock()

	select {
	case wrk.consensusStateChangedChannel <- true:
	default:
	}
}

// AddReceivedHeaderHandler adds a new handler function for a received header
func (wrk *Worker) AddReceivedHeaderHandler(handler func(data.HeaderHandler)) {
	wrk.mutReceivedHeadersHandler.Lock()
	wrk.receivedHeadersHandlers = append(wrk.receivedHeadersHandlers, handler)
	wrk.mutReceivedHeadersHandler.Unlock()
}

func (wrk *Worker) initReceivedMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages = wrk.consensusService.InitReceivedMessages()
	wrk.mutReceivedMessages.Unlock()
}

// AddReceivedMessageCall adds a new handler function for a received message type
func (wrk *Worker) AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls[messageType] = receivedMessageCall
	wrk.mutReceivedMessagesCalls.Unlock()
}

// RemoveAllReceivedMessagesCalls removes all the functions handlers
func (wrk *Worker) RemoveAllReceivedMessagesCalls() {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls = make(map[consensus.MessageType]func(context.Context, *consensus.Message) bool)
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *Worker) getCleanedList(cnsDataList []*consensus.Message) []*consensus.Message {
	cleanedCnsDataList := make([]*consensus.Message, 0)

	for i := 0; i < len(cnsDataList); i++ {
		if cnsDataList[i] == nil {
			continue
		}

		if wrk.roundHandler.Index() > cnsDataList[i].RoundIndex {
			continue
		}

		cleanedCnsDataList = append(cleanedCnsDataList, cnsDataList[i])
	}

	return cleanedCnsDataList
}

// ProcessReceivedMessage method redirects the received message to the channel which should handle it
func (wrk *Worker) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, _ p2p.MessageHandler) error {
	if check.IfNil(message) {
		return ErrNilMessage
	}
	if message.Data() == nil {
		return ErrNilDataToProcess
	}
	if len(message.Signature()) == 0 {
		return ErrNilSignatureOnP2PMessage
	}

	isPeerBlacklisted := wrk.peerBlacklistHandler.IsPeerBlacklisted(fromConnectedPeer)
	if isPeerBlacklisted {
		log.Debug("received message from blacklisted peer",
			"peer", fromConnectedPeer.Pretty(),
		)
		return ErrBlacklistedConsensusPeer
	}

	topic := GetConsensusTopicID(wrk.shardCoordinator)
	err := wrk.antifloodHandler.CanProcessMessagesOnTopic(message.Peer(), topic, 1, uint64(len(message.Data())), message.SeqNo())
	if err != nil {
		return err
	}

	defer func() {
		if wrk.shouldBlacklistPeer(err) {
			// this situation is so severe that we have to black list both the message originator and the connected peer
			// that disseminated this message.

			reason := fmt.Sprintf("blacklisted due to invalid consensus message: %s", err.Error())
			wrk.antifloodHandler.BlacklistPeer(message.Peer(), reason, common.InvalidMessageBlacklistDuration)
			wrk.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, common.InvalidMessageBlacklistDuration)
		}
	}()

	cnsMsg := &consensus.Message{}
	err = wrk.marshalizer.Unmarshal(cnsMsg, message.Data())
	if err != nil {
		return err
	}

	wrk.consensusState.ResetRoundsWithoutReceivedMessages(cnsMsg.GetPubKey(), message.Peer())

	if wrk.nodeRedundancyHandler.IsRedundancyNode() {
		wrk.nodeRedundancyHandler.ResetInactivityIfNeeded(
			wrk.consensusState.SelfPubKey(),
			string(cnsMsg.PubKey),
			message.Peer(),
		)
	}

	err = wrk.checkValidityAndProcessEquivalentMessages(cnsMsg, message)
	if err != nil {
		return err
	}

	wrk.networkShardingCollector.UpdatePeerIDInfo(message.Peer(), cnsMsg.PubKey, wrk.shardCoordinator.SelfId())

	msgType := consensus.MessageType(cnsMsg.MsgType)
	isMessageWithBlockBody := wrk.consensusService.IsMessageWithBlockBody(msgType)
	isMessageWithBlockHeader := wrk.consensusService.IsMessageWithBlockHeader(msgType)
	isMessageWithBlockBodyAndHeader := wrk.consensusService.IsMessageWithBlockBodyAndHeader(msgType)

	if isMessageWithBlockBody || isMessageWithBlockBodyAndHeader {
		wrk.doJobOnMessageWithBlockBody(cnsMsg)
	}

	if isMessageWithBlockHeader || isMessageWithBlockBodyAndHeader {
		err = wrk.doJobOnMessageWithHeader(cnsMsg)
		if err != nil {
			return err
		}
	}

	if wrk.consensusService.IsMessageWithSignature(msgType) {
		wrk.doJobOnMessageWithSignature(cnsMsg, message)
	}

	errNotCritical := wrk.checkSelfState(cnsMsg)
	if errNotCritical != nil {
		log.Trace("checkSelfState", "error", errNotCritical.Error())
		// in this case should return nil but do not process the message
		// nil error will mean that the interceptor will validate this message and broadcast it to the connected peers
		return nil
	}

	go wrk.executeReceivedMessages(cnsMsg)

	return nil
}

func (wrk *Worker) shouldBlacklistPeer(err error) bool {
	if err == nil ||
		errors.Is(err, ErrMessageForPastRound) ||
		errors.Is(err, ErrMessageForFutureRound) ||
		errors.Is(err, ErrNodeIsNotInEligibleList) ||
		errors.Is(err, errorsErd.ErrPIDMismatch) ||
		errors.Is(err, errorsErd.ErrSignatureMismatch) ||
		errors.Is(err, nodesCoordinator.ErrEpochNodesConfigDoesNotExist) ||
		errors.Is(err, ErrMessageTypeLimitReached) ||
		errors.Is(err, crypto.ErrAggSigNotValid) {
		return false
	}

	return true
}

func (wrk *Worker) doJobOnMessageWithBlockBody(cnsMsg *consensus.Message) {
	wrk.addBlockToPool(cnsMsg.GetBody())
}

func (wrk *Worker) doJobOnMessageWithHeader(cnsMsg *consensus.Message) error {
	headerHash := cnsMsg.BlockHeaderHash
	header := wrk.blockProcessor.DecodeBlockHeader(cnsMsg.Header)
	isHeaderInvalid := headerHash == nil || check.IfNil(header)
	if isHeaderInvalid {
		return fmt.Errorf("%w : received header from consensus topic is invalid",
			ErrInvalidHeader)
	}
	err := header.CheckFieldsForNil()
	if err != nil {
		return fmt.Errorf("%w : received header from consensus topic is invalid",
			err)
	}

	var valStatsRootHash []byte
	metaHeader, ok := header.(data.MetaHeaderHandler)
	if ok {
		valStatsRootHash = metaHeader.GetValidatorStatsRootHash()
	}

	log.Debug("received proposed block",
		"from", core.GetTrimmedPk(hex.EncodeToString(cnsMsg.PubKey)),
		"header hash", cnsMsg.BlockHeaderHash,
		"epoch", header.GetEpoch(),
		"round", header.GetRound(),
		"nonce", header.GetNonce(),
		"prev hash", header.GetPrevHash(),
		"nbTxs", header.GetTxCount(),
		"val stats root hash", valStatsRootHash)

	err = wrk.headerIntegrityVerifier.Verify(header)
	if err != nil {
		return fmt.Errorf("%w : verify header integrity from consensus topic failed", err)
	}

	err = wrk.headerSigVerifier.VerifyRandSeed(header)
	if err != nil {
		return fmt.Errorf("%w : verify rand seed for received header from consensus topic failed",
			err)
	}

	wrk.processReceivedHeaderMetric(cnsMsg)

	errNotCritical := wrk.forkDetector.AddHeader(header, headerHash, process.BHProposed, nil, nil)
	if errNotCritical != nil {
		log.Debug("add received header from consensus topic to fork detector failed",
			"error", errNotCritical.Error())
		// we should not return error here because the other peers connected to self might need this message
		// to advance the consensus
	}

	return nil
}

func (wrk *Worker) doJobOnMessageWithSignature(cnsMsg *consensus.Message, p2pMsg p2p.MessageP2P) {
	wrk.mutDisplayHashConsensusMessage.Lock()
	defer wrk.mutDisplayHashConsensusMessage.Unlock()

	hash := string(cnsMsg.BlockHeaderHash)
	wrk.mapDisplayHashConsensusMessage[hash] = append(wrk.mapDisplayHashConsensusMessage[hash], cnsMsg)

	wrk.consensusState.AddMessageWithSignature(string(cnsMsg.PubKey), p2pMsg)
}

func (wrk *Worker) addBlockToPool(bodyBytes []byte) {
	bodyHandler := wrk.blockProcessor.DecodeBlockBody(bodyBytes)
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return
	}

	for _, miniblock := range body.MiniBlocks {
		hash, err := core.CalculateHash(wrk.marshalizer, wrk.hasher, miniblock)
		if err != nil {
			return
		}
		wrk.poolAdder.Put(hash, miniblock, miniblock.Size())
	}
}

func (wrk *Worker) processReceivedHeaderMetric(cnsDta *consensus.Message) {
	if wrk.consensusState.ConsensusGroup() == nil || !wrk.consensusState.IsNodeLeaderInCurrentRound(string(cnsDta.PubKey)) {
		return
	}

	sinceRoundStart := time.Since(wrk.roundHandler.TimeStamp())
	if sinceRoundStart < 0 {
		sinceRoundStart = 0
	}
	percent := sinceRoundStart * 100 / wrk.roundHandler.TimeDuration()
	wrk.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlock, uint64(percent))
	wrk.appStatusHandler.SetStringValue(common.MetricRedundancyIsMainActive, strconv.FormatBool(wrk.nodeRedundancyHandler.IsMainMachineActive()))
}

func (wrk *Worker) checkSelfState(cnsDta *consensus.Message) error {
	isMultiKeyManagedBySelf := wrk.consensusState.keysHandler.IsKeyManagedByCurrentNode(cnsDta.PubKey)
	if wrk.consensusState.SelfPubKey() == string(cnsDta.PubKey) || isMultiKeyManagedBySelf {
		return ErrMessageFromItself
	}

	if wrk.consensusState.RoundCanceled && wrk.consensusState.RoundIndex == cnsDta.RoundIndex {
		return ErrRoundCanceled
	}

	return nil
}

func (wrk *Worker) executeReceivedMessages(cnsDta *consensus.Message) {
	wrk.mutReceivedMessages.Lock()

	msgType := consensus.MessageType(cnsDta.MsgType)
	cnsDataList := wrk.receivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.receivedMessages[msgType] = cnsDataList
	wrk.executeStoredMessages()

	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) executeStoredMessages() {
	for _, i := range wrk.consensusService.GetMessageRange() {
		cnsDataList := wrk.receivedMessages[i]
		if len(cnsDataList) == 0 {
			continue
		}
		wrk.executeMessage(cnsDataList)
		cleanedCnsDtaList := wrk.getCleanedList(cnsDataList)
		wrk.receivedMessages[i] = cleanedCnsDtaList
	}
}

func (wrk *Worker) executeMessage(cnsDtaList []*consensus.Message) {
	for i, cnsDta := range cnsDtaList {
		if cnsDta == nil {
			continue
		}
		if wrk.consensusState.RoundIndex != cnsDta.RoundIndex {
			continue
		}

		msgType := consensus.MessageType(cnsDta.MsgType)
		if !wrk.consensusService.CanProceed(wrk.consensusState, msgType) {
			continue
		}

		cnsDtaList[i] = nil
		select {
		case wrk.executeMessageChannel <- cnsDta:
		case <-wrk.closer.ChanClose():
			log.Debug("worker's executeMessage go routine is stopping...")
			return
		}
	}
}

// checkChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (wrk *Worker) checkChannels(ctx context.Context) {
	var rcvDta *consensus.Message

	for {
		select {
		case <-ctx.Done():
			log.Debug("worker's go routine is stopping...")
			return
		case rcvDta = <-wrk.executeMessageChannel:
		case <-time.After(sleepTime):
			continue
		}

		msgType := consensus.MessageType(rcvDta.MsgType)

		if callReceivedMessage, exist := wrk.receivedMessagesCalls[msgType]; exist {
			if callReceivedMessage(ctx, rcvDta) {
				select {
				case wrk.consensusStateChangedChannel <- true:
				default:
				}
			}
		}
	}
}

// Extend does an extension for the subround with subroundId
func (wrk *Worker) Extend(subroundId int) {
	wrk.consensusState.ExtendedCalled = true
	log.Debug("extend function is called",
		"subround", wrk.consensusService.GetSubroundName(subroundId))

	wrk.DisplayStatistics()

	if wrk.consensusService.IsSubroundStartRound(subroundId) {
		return
	}

	for wrk.consensusState.ProcessingBlock() {
		time.Sleep(time.Millisecond)
	}

	wrk.scheduledProcessor.ForceStopScheduledExecutionBlocking()
	wrk.blockProcessor.RevertCurrentBlock()
	log.Debug("current block is reverted")
}

// DisplayStatistics logs the consensus messages split on proposed headers
func (wrk *Worker) DisplayStatistics() {
	wrk.mutDisplayHashConsensusMessage.Lock()
	for hash, consensusMessages := range wrk.mapDisplayHashConsensusMessage {
		log.Debug("proposed header with signatures",
			"hash", []byte(hash),
			"sigs num", len(consensusMessages),
			"round", consensusMessages[0].RoundIndex,
		)

		for _, consensusMessage := range consensusMessages {
			log.Trace(core.GetTrimmedPk(hex.EncodeToString(consensusMessage.PubKey)))
		}

	}

	wrk.mapDisplayHashConsensusMessage = make(map[string][]*consensus.Message)

	wrk.mutDisplayHashConsensusMessage.Unlock()

	wrk.equivalentMessagesDebugger.DisplayEquivalentMessagesStatistics(wrk.getEquivalentMessages)
}

// GetConsensusStateChangedChannel gets the channel for the consensusStateChanged
func (wrk *Worker) GetConsensusStateChangedChannel() chan bool {
	return wrk.consensusStateChangedChannel
}

// ExecuteStoredMessages tries to execute all the messages received which are valid for execution
func (wrk *Worker) ExecuteStoredMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.executeStoredMessages()
	wrk.mutReceivedMessages.Unlock()
}

// Close will close the endless running go routine
func (wrk *Worker) Close() error {
	// calling close on the SafeCloser instance should be the last instruction called
	// (just to close some go routines started as edge cases that would otherwise hang)
	defer wrk.closer.Close()

	if wrk.cancelFunc != nil {
		wrk.cancelFunc()
	}

	wrk.cleanChannels()

	return nil
}

// ResetConsensusMessages resets at the start of each round all the previous consensus messages received
func (wrk *Worker) ResetConsensusMessages() {
	wrk.consensusMessageValidator.resetConsensusMessages()

	wrk.mutEquivalentMessages.Lock()
	wrk.equivalentMessages = make(map[string]*consensus.EquivalentMessageInfo)
	wrk.mutEquivalentMessages.Unlock()
}

func (wrk *Worker) checkValidityAndProcessEquivalentMessages(cnsMsg *consensus.Message, p2pMessage p2p.MessageP2P) error {
	msgType := consensus.MessageType(cnsMsg.MsgType)

	log.Trace("received message from consensus topic",
		"msg type", wrk.consensusService.GetStringValue(msgType),
		"from", cnsMsg.PubKey,
		"header hash", cnsMsg.BlockHeaderHash,
		"round", cnsMsg.RoundIndex,
		"size", len(p2pMessage.Data()),
	)

	if !wrk.shouldVerifyEquivalentMessages(msgType) {
		return wrk.consensusMessageValidator.checkConsensusMessageValidity(cnsMsg, p2pMessage.Peer())
	}

	wrk.mutEquivalentMessages.Lock()
	defer wrk.mutEquivalentMessages.Unlock()

	err := wrk.processEquivalentMessageUnprotected(cnsMsg)
	if err != nil {
		return err
	}

	err = wrk.consensusMessageValidator.checkConsensusMessageValidity(cnsMsg, p2pMessage.Peer())
	if err != nil {
		wrk.processInvalidEquivalentMessageUnprotected(cnsMsg.BlockHeaderHash)
		return err
	}

	return nil
}

func (wrk *Worker) shouldVerifyEquivalentMessages(msgType consensus.MessageType) bool {
	if !wrk.enableEpochsHandler.IsFlagEnabled(common.EquivalentMessagesFlag) {
		return false
	}

	return wrk.consensusService.IsMessageWithFinalInfo(msgType)
}

func (wrk *Worker) processEquivalentMessageUnprotected(cnsMsg *consensus.Message) error {
	hdrHash := string(cnsMsg.BlockHeaderHash)
	equivalentMsgInfo, ok := wrk.equivalentMessages[hdrHash]
	if !ok {
		equivalentMsgInfo = &consensus.EquivalentMessageInfo{}
		wrk.equivalentMessages[hdrHash] = equivalentMsgInfo
	}
	equivalentMsgInfo.NumMessages++

	if equivalentMsgInfo.Validated {
		return ErrEquivalentMessageAlreadyReceived
	}

	err := wrk.verifyEquivalentMessageSignature(cnsMsg)
	if err != nil {
		return err
	}

	equivalentMsgInfo.Validated = true

	return nil
}

func (wrk *Worker) verifyEquivalentMessageSignature(_ *consensus.Message) error {
	if check.IfNil(wrk.consensusState.Header) {
		return ErrNilHeader
	}

	header := wrk.consensusState.Header.ShallowClone()

	// TODO[Sorin]: after flag enabled, VerifySignature on previous hash, with the signature and bitmap from the proof on cnsMsg
	return wrk.headerSigVerifier.VerifySignature(header)
}

func (wrk *Worker) processInvalidEquivalentMessageUnprotected(blockHeaderHash []byte) {
	hdrHash := string(blockHeaderHash)
	delete(wrk.equivalentMessages, hdrHash)
}

// getEquivalentMessages returns a copy of the equivalent messages
func (wrk *Worker) getEquivalentMessages() map[string]*consensus.EquivalentMessageInfo {
	wrk.mutEquivalentMessages.RLock()
	defer wrk.mutEquivalentMessages.RUnlock()

	equivalentMessagesCopy := make(map[string]*consensus.EquivalentMessageInfo, len(wrk.equivalentMessages))
	for hash, cnt := range wrk.equivalentMessages {
		equivalentMessagesCopy[hash] = cnt
	}

	return equivalentMessagesCopy
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrk *Worker) IsInterfaceNil() bool {
	return wrk == nil
}

func (wrk *Worker) cleanChannels() {
	nrReads := core.EmptyChannel(wrk.consensusStateChangedChannel)
	log.Debug("close worker: emptied channel", "consensusStateChangedChannel nrReads", nrReads)

	nrReads = emptyChannel(wrk.executeMessageChannel)
	log.Debug("close worker: emptied channel", "executeMessageChannel nrReads", nrReads)
}

func emptyChannel(ch chan *consensus.Message) int {
	readsCnt := 0
	for {
		select {
		case <-ch:
			readsCnt++
		default:
			return readsCnt
		}
	}
}
