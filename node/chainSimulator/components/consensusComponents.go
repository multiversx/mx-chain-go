package components

import (
	"errors"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-storage-go/timecache"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/blacklist"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls/proxy"
	"github.com/multiversx/mx-chain-go/consensus/spos/sposFactory"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
)

var (
	errNodeNotInConsensusMode   = errors.New("node was not built in consensus-path execution mode")
	errConsensusDriveClosed     = errors.New("consensus drive was closed")
	errRoundHandlerNotRearmable = errors.New("consensus round handler cannot rearm the current round")
)

const numSignatureGoRoutinesThrottler = 30

// advanceableSyncTimer is the manual-clock surface the consensus drive needs from the node's
// sync timer: only NewManualSyncTimer (installed in consensus mode) satisfies it.
type advanceableSyncTimer interface {
	AdvanceTime(d time.Duration) time.Time
}

// syncedBootstrapper is a no-op bootstrapper that always reports a synchronized node. The SPoS
// start-round subround refuses to proceed unless the bootstrapper reports NsSynchronized, and a
// simulator node with a single validator per shard has nothing to sync from — it produces its own
// blocks — so the real syncing bootstrapper (and its background goroutine) is replaced by this.
type syncedBootstrapper struct{}

// AddSyncStateListener does nothing, this bootstrapper never changes sync state
func (b *syncedBootstrapper) AddSyncStateListener(_ func(isSyncing bool)) {}

// GetNodeState always reports a synchronized node so the start-round subround can proceed
func (b *syncedBootstrapper) GetNodeState() common.NodeState {
	return common.NsSynchronized
}

// StartSyncingBlocks does nothing, there is no background syncing in the simulator
func (b *syncedBootstrapper) StartSyncingBlocks() error {
	return nil
}

// Close does nothing
func (b *syncedBootstrapper) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *syncedBootstrapper) IsInterfaceNil() bool {
	return b == nil
}

// nodeConsensusDrive owns the manual seam that lets the simulator step one node's chronology and
// SPoS subrounds forward one round at a time.
//
// The chronology goroutine loops: receive one tick, run exactly one subround, acknowledge its
// completion, repeat. The simulator sends a tick to every validator in a shard before waiting for
// any acknowledgement. Validators therefore execute the same subround concurrently, while a fast
// validator cannot overtake a slow validator. The single chronology goroutine is reused across
// consensus-version switches and also acknowledges ticks drained while temporarily paused.
type nodeConsensusDrive struct {
	syncTimer      advanceableSyncTimer
	tickChan       chan time.Time
	stepDoneChan   chan struct{}
	closeChan      chan struct{}
	closeOnce      sync.Once
	roundHandler   consensusRoundHandler
	consensusState *spos.ConsensusState
	chronology     *simChronology
}

// roundDurationProvider is the slice of the round handler the drive needs to advance the clock by
// exactly one round: TimeDuration() follows the Supernova transition (it returns the Supernova
// round duration once the activation round is reached), so the manual clock keeps advancing one
// round at a time across the duration switch. A fixed construction-time duration would jump
// roundDuration/supernovaRoundDuration rounds per advance after the transition, silently skipping
// the rounds in between.
type consensusRoundHandler interface {
	TimeDuration() time.Duration
	RevertOneRound()
}

type consensusParticipationRoundHandler interface {
	setConsensusParticipant(isParticipant bool)
	resetConsensusParticipation()
	prepareConsensusStep()
}

// advanceClock bumps the node's manual clock by one round duration — the round handler's CURRENT
// one, so the advance stays one round across the Supernova duration switch. Called once per round
// per node.
func (d *nodeConsensusDrive) advanceClock() {
	d.resetConsensusParticipation()
	d.syncTimer.AdvanceTime(d.roundHandler.TimeDuration())
}

// rearmCurrentRound makes the next chronology tick initialize the current manual-clock round.
// This is needed only when the simulator retries a failed round without advancing time.
func (d *nodeConsensusDrive) rearmCurrentRound() {
	d.resetConsensusParticipation()
	d.roundHandler.RevertOneRound()
}

// step fires one chronology tick (one subround). The closeChan escape keeps step from blocking
// forever if the chronology goroutine is gone during teardown.
func (d *nodeConsensusDrive) step() error {
	if handler, ok := d.roundHandler.(consensusParticipationRoundHandler); ok {
		handler.prepareConsensusStep()
	}

	select {
	case d.tickChan <- time.Time{}:
	case <-d.closeChan:
		return errConsensusDriveClosed
	}

	return nil
}

// waitStep waits for the chronology to complete the subround started by step.
func (d *nodeConsensusDrive) waitStep() error {
	select {
	case <-d.stepDoneChan:
		d.updateConsensusParticipation()
	case <-d.closeChan:
		return errConsensusDriveClosed
	}

	return nil
}

// state returns the chronology position and its restart generation. A consensus implementation
// switch restarts the chronology at START_ROUND while a peer can already be farther through the
// same logical round; the simulator uses this state to let the restarted peer catch up without
// advancing the peer that is already ahead.
func (d *nodeConsensusDrive) state() (int, uint64) {
	return d.chronology.getSubroundId(), d.chronology.getGeneration()
}

func (d *nodeConsensusDrive) updateConsensusParticipation() {
	handler, ok := d.roundHandler.(consensusParticipationRoundHandler)
	if !ok {
		return
	}

	isParticipant := d.consensusState.IsNodeInConsensusGroup(d.consensusState.SelfPubKey()) ||
		d.consensusState.IsMultiKeyInConsensusGroup()
	handler.setConsensusParticipant(isParticipant)
}

func (d *nodeConsensusDrive) resetConsensusParticipation() {
	handler, ok := d.roundHandler.(consensusParticipationRoundHandler)
	if ok {
		handler.resetConsensusParticipation()
	}
}

// Close implements the closer contract used by the node's close handler
func (d *nodeConsensusDrive) Close() error {
	d.closeOnce.Do(func() {
		close(d.closeChan)
	})
	return nil
}

// createConsensusComponents assembles the real consensus stack (chronology, worker, SPoS subrounds)
// over the node's already-built components and returns a drive the simulator can step manually. It
// mirrors the production consensus components factory's Create() flow, with two simulator-specific
// substitutions: the chronology is the tick-driven simChronology instead of the real time-driven
// one, and the bootstrapper is the synced no-op above instead of the real syncing one.
func (node *testOnlyProcessingNode) createConsensusComponents(generalConfig config.Config) (*nodeConsensusDrive, error) {
	syncTimer, ok := node.CoreComponentsHolder.SyncTimer().(advanceableSyncTimer)
	if !ok {
		return nil, errNodeNotInConsensusMode
	}

	tickChan := make(chan time.Time)
	stepDoneChan := make(chan struct{})
	closeChan := make(chan struct{})
	chronologyHandler, err := createConsensusChronology(node, tickChan, stepDoneChan)
	if err != nil {
		return nil, err
	}

	epoch := node.consensusStartEpoch()
	consensusGroupSize, err := consensusGroupSizeForEpoch(node, epoch)
	if err != nil {
		return nil, err
	}

	consensusState, err := node.createConsensusState(epoch, consensusGroupSize)
	if err != nil {
		return nil, err
	}

	consensusService, err := sposFactory.GetConsensusCoreFactory(generalConfig.Consensus.Type)
	if err != nil {
		return nil, err
	}

	bootstrapper := &syncedBootstrapper{}

	scheduledProcessor, err := spos.NewScheduledProcessorWrapper(spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                node.CoreComponentsHolder.SyncTimer(),
		Processor:                node.ProcessComponentsHolder.BlockProcessor(),
		RoundTimeDurationHandler: node.CoreComponentsHolder.RoundHandler(),
	})
	if err != nil {
		return nil, err
	}

	var p2pSigningHandler consensus.P2PSigningHandler
	p2pSigningHandler, err = p2pFactory.NewMessageVerifier(p2pFactory.ArgsMessageVerifier{
		Marshaller: node.CoreComponentsHolder.InternalMarshalizer(),
		P2PSigner:  node.NetworkComponentsHolder.NetworkMessenger(),
		Logger:     logger.GetOrCreate("main/p2p/messagecheck"),
	})
	if err != nil {
		return nil, err
	}

	invalidSignersCache, err := spos.NewInvalidSignersCache(spos.ArgInvalidSignersCache{
		Hasher:         node.CoreComponentsHolder.Hasher(),
		SigningHandler: p2pSigningHandler,
		Marshaller:     node.CoreComponentsHolder.InternalMarshalizer(),
	})
	if err != nil {
		return nil, err
	}

	peerBlacklistHandler, err := createConsensusPeerBlacklistHandler()
	if err != nil {
		return nil, err
	}

	peerSignatureHandler := node.CryptoComponentsHolder.PeerSignatureHandler()
	if provider, ok := node.syncedBroadcastNetwork.(interface {
		wrapPeerSignatureHandler(crypto.PeerSignatureHandler) crypto.PeerSignatureHandler
	}); ok {
		peerSignatureHandler = provider.wrapPeerSignatureHandler(peerSignatureHandler)
	}

	baseWorker, err := node.createConsensusWorker(createConsensusWorkerArgs{
		consensusService:     consensusService,
		bootstrapper:         bootstrapper,
		consensusState:       consensusState,
		scheduledProcessor:   scheduledProcessor,
		peerSignatureHandler: peerSignatureHandler,
		peerBlacklistHandler: peerBlacklistHandler,
		invalidSignersCache:  invalidSignersCache,
		generalConfig:        generalConfig,
	})
	if err != nil {
		return nil, err
	}

	deliveryTracker, _ := node.syncedBroadcastNetwork.(blockBodyDeliveryTracker)
	headerDeliveryTracker, _ := node.syncedBroadcastNetwork.(blockHeaderDeliveryTracker)
	worker := newTrackedBlockBodyWorker(
		baseWorker,
		node.CoreComponentsHolder.InternalMarshalizer(),
		consensusState,
		deliveryTracker,
		headerDeliveryTracker,
		node.ProcessComponentsHolder.ShardCoordinator().SelfId(),
	)

	worker.StartWorking()
	node.DataComponentsHolder.Datapool().Headers().RegisterHandler(worker.ReceivedHeader)
	node.NetworkComponentsHolder.InputAntiFloodHandler().SetConsensusSizeNotifier(
		node.CoreComponentsHolder.ChainParametersSubscriber(),
		node.ProcessComponentsHolder.ShardCoordinator().SelfId(),
	)

	err = node.createConsensusTopic(worker)
	if err != nil {
		return nil, err
	}

	consensusDataContainer, err := spos.NewConsensusCore(node.consensusCoreArgs(createConsensusCoreArgs{
		consensusService:     consensusService,
		bootstrapper:         bootstrapper,
		chronology:           chronologyHandler,
		scheduledProcessor:   scheduledProcessor,
		p2pSigningHandler:    p2pSigningHandler,
		peerSignatureHandler: peerSignatureHandler,
		peerBlacklistHandler: peerBlacklistHandler,
		invalidSignersCache:  invalidSignersCache,
	}))
	if err != nil {
		return nil, err
	}

	signatureThrottler, err := throttler.NewNumGoRoutinesThrottler(numSignatureGoRoutinesThrottler)
	if err != nil {
		return nil, err
	}

	subroundsHandler, err := proxy.NewSubroundsHandler(&proxy.SubroundsHandlerArgs{
		Chronology:           chronologyHandler,
		ConsensusCoreHandler: consensusDataContainer,
		ConsensusState:       consensusState,
		Worker:               worker,
		SignatureThrottler:   signatureThrottler,
		AppStatusHandler:     node.StatusCoreComponents.AppStatusHandler(),
		OutportHandler:       node.StatusComponentsHolder.OutportHandler(),
		SentSignatureTracker: node.ProcessComponentsHolder.SentSignaturesTracker(),
		EnableEpochsHandler:  node.CoreComponentsHolder.EnableEpochsHandler(),
		ChainID:              []byte(node.CoreComponentsHolder.ChainID()),
		CurrentPid:           node.NetworkComponentsHolder.NetworkMessenger().ID(),
	})
	if err != nil {
		return nil, err
	}

	// Start launches the chronology rounds goroutine, which immediately blocks on the injected
	// tick channel: nothing advances until the simulator calls advance()
	err = subroundsHandler.Start(epoch)
	if err != nil {
		return nil, err
	}

	// the chronology's interface Close only pauses tick processing (the proxy restarts the
	// chronology on every consensus-version switch and must keep its single goroutine alive); the
	// dedicated stopper cancels the goroutine's context at node teardown, and the drive's Close
	// additionally releases closeChan so advance() can never block forever on a tick send if the
	// goroutine has gone away
	node.closeHandler.AddComponent(&chronologyStopper{chronology: chronologyHandler})
	node.closeHandler.AddComponent(worker)
	// the peer blacklist starts a background time-cache sweeper goroutine in its constructor; register
	// it so node Close cancels that goroutine instead of leaking one sweeper per consensus node — over a
	// long test suite the accumulated sweepers starve the manual drive's timing-sensitive convergence
	node.closeHandler.AddComponent(peerBlacklistHandler)

	roundHandler, ok := node.CoreComponentsHolder.RoundHandler().(consensusRoundHandler)
	if !ok {
		return nil, errRoundHandlerNotRearmable
	}
	boundedRoundHandler, ok := roundHandler.(*boundedWaitRoundHandler)
	if !ok {
		return nil, errRoundHandlerNotRearmable
	}
	simMessenger, ok := node.NetworkComponentsHolder.NetworkMessenger().(*syncedMessenger)
	if !ok {
		return nil, errNodeNotInConsensusMode
	}
	simMessenger.setConsensusMessageFilter(boundedRoundHandler.shouldReceiveConsensusMessage)

	drive := &nodeConsensusDrive{
		syncTimer:      syncTimer,
		tickChan:       tickChan,
		stepDoneChan:   stepDoneChan,
		closeChan:      closeChan,
		roundHandler:   roundHandler,
		consensusState: consensusState,
		chronology:     chronologyHandler,
	}
	node.closeHandler.AddComponent(drive)

	return drive, nil
}

// chronologyStopper stops the simChronology goroutine at node teardown. The chronology's interface
// Close only pauses the loop (so the proxy's epoch-switch restart keeps the one goroutine alive), so
// teardown needs this explicit stop to cancel the goroutine's context and avoid leaking it.
type chronologyStopper struct {
	chronology *simChronology
}

// Close stops the underlying chronology goroutine
func (c *chronologyStopper) Close() error {
	c.chronology.stop()
	return nil
}

func createConsensusChronology(
	node *testOnlyProcessingNode,
	tickChan chan time.Time,
	stepDoneChan chan struct{},
) (*simChronology, error) {
	return NewSimChronology(ArgsSimChronology{
		GenesisTime:         node.CoreComponentsHolder.GenesisTime(),
		RoundHandler:        node.CoreComponentsHolder.RoundHandler(),
		SyncTimer:           node.CoreComponentsHolder.SyncTimer(),
		AppStatusHandler:    node.StatusCoreComponents.AppStatusHandler(),
		EnableEpochsHandler: node.CoreComponentsHolder.EnableEpochsHandler(),
		EnableRoundsHandler: node.CoreComponentsHolder.EnableRoundsHandler(),
		ConfigsHandler:      node.CoreComponentsHolder.CommonConfigsHandler(),
		TickChan:            tickChan,
		StepDoneChan:        stepDoneChan,
	})
}

// consensusStartEpoch mirrors the factory's getEpoch: the current block's epoch, or the genesis
// epoch before any block exists
func (node *testOnlyProcessingNode) consensusStartEpoch() uint32 {
	blockchain := node.DataComponentsHolder.Blockchain()
	epoch := blockchain.GetGenesisHeader().GetEpoch()
	currentHeader := blockchain.GetCurrentBlockHeader()
	if !check.IfNil(currentHeader) {
		epoch = currentHeader.GetEpoch()
	}

	return epoch
}

func (node *testOnlyProcessingNode) createConsensusState(epoch uint32, consensusGroupSize int) (*spos.ConsensusState, error) {
	selfId, err := node.CryptoComponentsHolder.PublicKey().ToByteArray()
	if err != nil {
		return nil, err
	}

	eligibleNodesPubKeys, err := node.NodesCoordinator.GetConsensusWhitelistedNodes(epoch)
	if err != nil {
		return nil, err
	}

	roundConsensus, err := spos.NewRoundConsensus(
		eligibleNodesPubKeys,
		consensusGroupSize,
		string(selfId),
		node.CryptoComponentsHolder.KeysHandler(),
	)
	if err != nil {
		return nil, err
	}

	roundConsensus.ResetRoundState()

	roundThreshold := spos.NewRoundThreshold()
	roundStatus := spos.NewRoundStatus()
	roundStatus.ResetRoundStatus()

	return spos.NewConsensusState(
		roundConsensus,
		roundThreshold,
		roundStatus,
		node.ProcessComponentsHolder.NodeRedundancyHandler(),
	), nil
}

func (node *testOnlyProcessingNode) createConsensusTopic(worker spos.WorkerHandler) error {
	shardCoordinator := node.ProcessComponentsHolder.ShardCoordinator()
	messenger := node.NetworkComponentsHolder.NetworkMessenger()

	consensusTopic := common.ConsensusTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())
	if !messenger.HasTopic(consensusTopic) {
		err := messenger.CreateTopic(consensusTopic, true)
		if err != nil {
			return err
		}
	}

	return messenger.RegisterMessageProcessor(consensusTopic, common.DefaultInterceptorsIdentifier, worker)
}

func createConsensusPeerBlacklistHandler() (consensus.PeerBlacklistHandler, error) {
	cache := timecache.NewTimeCache(300 * time.Second)
	peerCacher, err := timecache.NewPeerTimeCache(cache)
	if err != nil {
		return nil, err
	}

	return blacklist.NewPeerBlacklist(blacklist.PeerBlackListArgs{PeerCacher: peerCacher})
}

// consensusGroupSizeForEpoch mirrors the factory's getConsensusGroupSize helper
func consensusGroupSizeForEpoch(node *testOnlyProcessingNode, epoch uint32) (int, error) {
	shardCoordinator := node.ProcessComponentsHolder.ShardCoordinator()
	consensusGroupSize := node.NodesCoordinator.ConsensusGroupSizeForShardAndEpoch(shardCoordinator.SelfId(), epoch)
	if consensusGroupSize > 0 {
		return consensusGroupSize, nil
	}

	nodesConfig := node.CoreComponentsHolder.GenesisNodesSetup()
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return int(nodesConfig.GetMetaConsensusGroupSize()), nil
	}

	return int(nodesConfig.GetShardConsensusGroupSize()), nil
}

type createConsensusWorkerArgs struct {
	consensusService     spos.ConsensusService
	bootstrapper         *syncedBootstrapper
	consensusState       *spos.ConsensusState
	scheduledProcessor   consensus.ScheduledProcessor
	peerSignatureHandler crypto.PeerSignatureHandler
	peerBlacklistHandler consensus.PeerBlacklistHandler
	invalidSignersCache  spos.InvalidSignersCache
	generalConfig        config.Config
}

func (node *testOnlyProcessingNode) createConsensusWorker(args createConsensusWorkerArgs) (*spos.Worker, error) {
	marshalizer := node.CoreComponentsHolder.InternalMarshalizer()
	sizeCheckDelta := args.generalConfig.Marshalizer.SizeCheckDelta
	if sizeCheckDelta > 0 {
		marshalizer = marshal.NewSizeCheckUnmarshalizer(marshalizer, sizeCheckDelta)
	}

	return spos.NewWorker(&spos.WorkerArgs{
		ConsensusService:         args.consensusService,
		BlockChain:               node.DataComponentsHolder.Blockchain(),
		BlockProcessor:           node.ProcessComponentsHolder.BlockProcessor(),
		ScheduledProcessor:       args.scheduledProcessor,
		Bootstrapper:             args.bootstrapper,
		BroadcastMessenger:       node.GetBroadcastMessenger(),
		ConsensusState:           args.consensusState,
		ForkDetector:             node.ProcessComponentsHolder.ForkDetector(),
		PeerSignatureHandler:     args.peerSignatureHandler,
		Marshalizer:              marshalizer,
		Hasher:                   node.CoreComponentsHolder.Hasher(),
		RoundHandler:             node.CoreComponentsHolder.RoundHandler(),
		ShardCoordinator:         node.ProcessComponentsHolder.ShardCoordinator(),
		SyncTimer:                node.CoreComponentsHolder.SyncTimer(),
		HeaderSigVerifier:        node.ProcessComponentsHolder.HeaderSigVerifier(),
		HeaderIntegrityVerifier:  node.ProcessComponentsHolder.HeaderIntegrityVerifier(),
		ChainID:                  []byte(node.CoreComponentsHolder.ChainID()),
		NetworkShardingCollector: node.ProcessComponentsHolder.PeerShardMapper(),
		AntifloodHandler:         node.NetworkComponentsHolder.InputAntiFloodHandler(),
		PoolAdder:                node.DataComponentsHolder.Datapool().MiniBlocks(),
		WhiteListHandler:         node.ProcessComponentsHolder.WhiteListHandler(),
		SignatureSize:            args.generalConfig.ValidatorPubkeyConverter.SignatureLength,
		PublicKeySize:            args.generalConfig.ValidatorPubkeyConverter.Length,
		AppStatusHandler:         node.StatusCoreComponents.AppStatusHandler(),
		NodeRedundancyHandler:    node.ProcessComponentsHolder.NodeRedundancyHandler(),
		PeerBlacklistHandler:     args.peerBlacklistHandler,
		EnableEpochsHandler:      node.CoreComponentsHolder.EnableEpochsHandler(),
		InvalidSignersCache:      args.invalidSignersCache,
		EnableRoundsHandler:      node.CoreComponentsHolder.EnableRoundsHandler(),
	})
}

type createConsensusCoreArgs struct {
	consensusService     spos.ConsensusService
	bootstrapper         *syncedBootstrapper
	chronology           consensus.ChronologyHandler
	scheduledProcessor   consensus.ScheduledProcessor
	p2pSigningHandler    consensus.P2PSigningHandler
	peerSignatureHandler crypto.PeerSignatureHandler
	peerBlacklistHandler consensus.PeerBlacklistHandler
	invalidSignersCache  spos.InvalidSignersCache
}

func (node *testOnlyProcessingNode) consensusCoreArgs(args createConsensusCoreArgs) *spos.ConsensusCoreArgs {
	return &spos.ConsensusCoreArgs{
		BlockChain:                    node.DataComponentsHolder.Blockchain(),
		BlockProcessor:                node.ProcessComponentsHolder.BlockProcessor(),
		ExecutionManager:              node.ProcessComponentsHolder.ExecutionManager(),
		Bootstrapper:                  args.bootstrapper,
		BroadcastMessenger:            node.GetBroadcastMessenger(),
		ChronologyHandler:             args.chronology,
		Hasher:                        node.CoreComponentsHolder.Hasher(),
		Marshalizer:                   node.CoreComponentsHolder.InternalMarshalizer(),
		MultiSignerContainer:          node.CryptoComponentsHolder.MultiSignerContainer(),
		RoundHandler:                  node.CoreComponentsHolder.RoundHandler(),
		ShardCoordinator:              node.ProcessComponentsHolder.ShardCoordinator(),
		NodesCoordinator:              node.NodesCoordinator,
		SyncTimer:                     node.CoreComponentsHolder.SyncTimer(),
		EpochStartRegistrationHandler: node.ProcessComponentsHolder.EpochStartNotifier(),
		AntifloodHandler:              node.NetworkComponentsHolder.InputAntiFloodHandler(),
		PeerHonestyHandler:            node.NetworkComponentsHolder.PeerHonestyHandler(),
		HeaderSigVerifier:             node.ProcessComponentsHolder.HeaderSigVerifier(),
		FallbackHeaderValidator:       node.ProcessComponentsHolder.FallbackHeaderValidator(),
		NodeRedundancyHandler:         node.ProcessComponentsHolder.NodeRedundancyHandler(),
		ScheduledProcessor:            args.scheduledProcessor,
		MessageSigningHandler:         args.p2pSigningHandler,
		PeerBlacklistHandler:          args.peerBlacklistHandler,
		PeerSignatureHandler:          args.peerSignatureHandler,
		SigningHandler:                node.CryptoComponentsHolder.ConsensusSigningHandler(),
		EnableEpochsHandler:           node.CoreComponentsHolder.EnableEpochsHandler(),
		EnableRoundsHandler:           node.CoreComponentsHolder.EnableRoundsHandler(),
		EquivalentProofsPool:          node.DataComponentsHolder.Datapool().Proofs(),
		EpochNotifier:                 node.CoreComponentsHolder.EpochNotifier(),
		InvalidSignersCache:           args.invalidSignersCache,
		MessagesHandler:               args.consensusService,
		AOTSelector:                   node.ProcessComponentsHolder.AOTSelector(),
		CommonConfigsHandler:          node.CoreComponentsHolder.CommonConfigsHandler(),
	}
}
