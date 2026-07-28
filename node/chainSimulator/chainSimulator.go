package chainSimulator

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/heartbeat"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorErrors "github.com/multiversx/mx-chain-go/node/chainSimulator/errors"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

const delaySendTxs = time.Millisecond

var log = logger.GetOrCreate("chainSimulator")

type transactionWithResult struct {
	hexHash string
	tx      *transaction.Transaction
	result  *transaction.ApiTransactionResult
}

// ConsensusMode selects how the chain simulator produces blocks.
type ConsensusMode uint8

const (
	// ConsensusModeDisabled keeps the historical direct block-production path.
	ConsensusModeDisabled ConsensusMode = iota
	// ConsensusModeBLS runs the real SPoS consensus path with BLS cryptography.
	ConsensusModeBLS
	// ConsensusModeFastCrypto runs the real SPoS consensus path with deterministic simulator-only cryptography.
	ConsensusModeFastCrypto
)

func (mode ConsensusMode) isValid() bool {
	return mode <= ConsensusModeFastCrypto
}

func (mode ConsensusMode) isEnabled() bool {
	return mode == ConsensusModeBLS || mode == ConsensusModeFastCrypto
}

func (mode ConsensusMode) isFastCrypto() bool {
	return mode == ConsensusModeFastCrypto
}

// ArgsChainSimulator holds the arguments needed to create a new instance of simulator
type ArgsChainSimulator struct {
	BypassTxSignatureCheck         bool
	BypassBlockSignatureCheck      bool
	TempDir                        string
	PathToInitialConfig            string
	NumOfShards                    uint32
	MinNodesPerShard               uint32
	MetaChainMinNodes              uint32
	Hysteresis                     float32
	NumNodesWaitingListShard       uint32
	NumNodesWaitingListMeta        uint32
	GenesisTimestamp               int64
	InitialRound                   int64
	InitialEpoch                   uint32
	InitialNonce                   uint64
	RoundDurationInMillis          uint64
	SupernovaRoundDurationInMillis uint64
	RoundsPerEpoch                 core.OptionalUint64
	SupernovaRoundsPerEpoch        core.OptionalUint64
	ApiInterface                   components.APIConfigurator
	AlterConfigsFunction           func(cfg *config.Configs)
	VmQueryDelayAfterStartInMs     uint64
	CreateBlockMaxTimePercent      float64
	BypassCreateBlockTimeCheck     bool
	// ConsensusMode selects direct block production, real SPoS consensus with BLS, or real SPoS
	// consensus with deterministic simulator-only cryptography.
	ConsensusMode ConsensusMode
}

// ArgsBaseChainSimulator holds the arguments needed to create a new instance of simulator
type ArgsBaseChainSimulator struct {
	ArgsChainSimulator
	ConsensusGroupSize          uint32
	MetaChainConsensusGroupSize uint32
}

type shardChainHandler struct {
	shardID uint32
	handler ChainHandler
}

type simulator struct {
	chanStopNodeProcess    chan endProcess.ArgEndProcess
	syncedBroadcastNetwork components.SyncedBroadcastNetworkHandler
	handlers               []shardChainHandler
	initialWalletKeys      *dtos.InitialWalletKeys
	initialStakedKeys      map[string]*dtos.BLSKey
	validatorsPrivateKeys  []crypto.PrivateKey
	nodes                  map[uint32]process.NodeHandler
	// consensusNodes holds every eligible and waiting validator node per shard. In direct mode it
	// has a single entry per shard, identical to nodes[shardID]. In consensus mode all single-key
	// nodes are driven each round; waiting nodes are kept caught up so they can become eligible.
	consensusNodes map[uint32][]process.NodeHandler
	numOfShards    uint32
	// genesisTime is the chain's genesis instant, shared by every node. Direct mode keeps the
	// historical behavior (time.Now() at construction); consensus runs use the caller-owned
	// GenesisTimestamp so every validator derives the same round.
	genesisTime time.Time
	// enableConsensus selects consensus-path execution: produceRound drives the real
	// chronology/SPoS state machine instead of the direct blocksCreator
	enableConsensus          bool
	deliveredConsensusProofs map[string]struct{}
	mutex                    sync.RWMutex
}

// NewChainSimulator will create a new instance of simulator
func NewChainSimulator(args ArgsChainSimulator) (*simulator, error) {
	return NewBaseChainSimulator(ArgsBaseChainSimulator{
		ArgsChainSimulator:          args,
		ConsensusGroupSize:          args.MinNodesPerShard,
		MetaChainConsensusGroupSize: args.MetaChainMinNodes,
	})
}

// NewBaseChainSimulator will create a new instance of simulator
func NewBaseChainSimulator(args ArgsBaseChainSimulator) (*simulator, error) {
	if !args.ConsensusMode.isValid() {
		return nil, chainSimulatorErrors.ErrInvalidConsensusMode
	}

	instance := &simulator{
		syncedBroadcastNetwork:   components.NewSyncedBroadcastNetwork(),
		nodes:                    make(map[uint32]process.NodeHandler),
		consensusNodes:           make(map[uint32][]process.NodeHandler),
		handlers:                 make([]shardChainHandler, 0, args.NumOfShards+1),
		numOfShards:              args.NumOfShards,
		chanStopNodeProcess:      make(chan endProcess.ArgEndProcess),
		mutex:                    sync.RWMutex{},
		initialStakedKeys:        make(map[string]*dtos.BLSKey),
		enableConsensus:          args.ConsensusMode.isEnabled(),
		deliveredConsensusProofs: make(map[string]struct{}),
	}

	err := instance.createChainHandlers(args)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func consensusAwareAlterConfigs(args ArgsBaseChainSimulator) func(cfg *config.Configs) {
	userAlterConfigs := args.AlterConfigsFunction
	if !args.ConsensusMode.isEnabled() {
		return userAlterConfigs
	}

	return func(cfg *config.Configs) {
		if userAlterConfigs != nil {
			userAlterConfigs(cfg)
		}

		const minimumConsensusHistory = 2000
		if cfg.GeneralConfig.ProofsPoolConfig.CleanupNonceDelta < minimumConsensusHistory {
			cfg.GeneralConfig.ProofsPoolConfig.CleanupNonceDelta = minimumConsensusHistory
		}
		if cfg.GeneralConfig.HeadersPoolConfig.MaxHeadersPerShard < minimumConsensusHistory {
			cfg.GeneralConfig.HeadersPoolConfig.MaxHeadersPerShard = minimumConsensusHistory
		}
	}
}

// simulatorHeartbeatMonitor is what the construction-time heartbeat monitor satisfies: both
// the node factory's view and the blocks creator's view
type simulatorHeartbeatMonitor interface {
	factory.HeartbeatV2Monitor
	process.HeartbeatMonitorWithSet
}

func (s *simulator) createChainHandlers(args ArgsBaseChainSimulator) error {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:                    args.NumOfShards,
		OriginalConfigsPath:            args.PathToInitialConfig,
		RoundDurationInMillis:          args.RoundDurationInMillis,
		SupernovaRoundDurationInMillis: args.SupernovaRoundDurationInMillis,
		TempDir:                        args.TempDir,
		MinNodesPerShard:               args.MinNodesPerShard,
		ConsensusGroupSize:             args.ConsensusGroupSize,
		MetaChainMinNodes:              args.MetaChainMinNodes,
		MetaChainConsensusGroupSize:    args.MetaChainConsensusGroupSize,
		Hysteresis:                     args.Hysteresis,
		RoundsPerEpoch:                 args.RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        args.SupernovaRoundsPerEpoch,
		InitialEpoch:                   args.InitialEpoch,
		AlterConfigsFunction:           consensusAwareAlterConfigs(args),
		NumNodesWaitingListShard:       args.NumNodesWaitingListShard,
		NumNodesWaitingListMeta:        args.NumNodesWaitingListMeta,
		InitialRound:                   args.InitialRound,
	})
	if err != nil {
		return err
	}

	// Direct mode keeps the historical wall-clock genesis. Consensus nodes share the caller-owned
	// genesis timestamp so their manual clocks derive the same round.
	s.genesisTime = time.Now()
	if args.ConsensusMode.isEnabled() {
		s.genesisTime = time.Unix(args.GenesisTimestamp, 0)
	}

	monitor := heartbeat.NewHeartbeatMonitor()

	err = s.createEligibleNodes(*outputConfigs, args, monitor)
	if err != nil {
		return err
	}

	err = s.createWaitingNodes(*outputConfigs, args, monitor)
	if err != nil {
		return err
	}

	s.initialWalletKeys = outputConfigs.InitialWallets
	s.validatorsPrivateKeys = outputConfigs.ValidatorsPrivateKeys

	s.addProofs()
	s.setBasePeerIds()

	log.Info("running the chain simulator with the following parameters",
		"number of shards (including meta)", args.NumOfShards+1,
		"original config path", args.PathToInitialConfig,
		"temporary path", args.TempDir)

	return nil
}

func (s *simulator) setBasePeerIds() {
	peerIds := make(map[uint32]core.PeerID, 0)
	for _, nodeHandler := range s.nodes {
		peerID := nodeHandler.GetNetworkComponents().NetworkMessenger().ID()
		peerIds[nodeHandler.GetShardCoordinator().SelfId()] = peerID
	}

	for _, nodes := range s.consensusNodes {
		for _, nodeHandler := range nodes {
			nodeHandler.SetBasePeers(peerIds)
		}
	}
}

func (s *simulator) addProofs() {
	proofs := make([]*block.HeaderProof, 0, len(s.nodes))

	for shardID, nodeHandler := range s.nodes {
		genesisHeader := nodeHandler.GetChainHandler().GetGenesisHeader()
		hash := nodeHandler.GetChainHandler().GetGenesisHeaderHash()
		proofs = append(proofs, &block.HeaderProof{
			HeaderHash:     hash,
			HeaderEpoch:    genesisHeader.GetEpoch(),
			HeaderNonce:    genesisHeader.GetNonce(),
			HeaderShardId:  shardID,
			HeaderRound:    genesisHeader.GetRound(),
			IsStartOfEpoch: false,
		})
	}

	// every node needs the genesis proofs it will reason about: all metachain nodes notarize all
	// shards' genesis, and each shard's nodes need their own shard's genesis proof
	for _, proof := range proofs {
		for _, metaNode := range s.consensusNodes[core.MetachainShardId] {
			_ = metaNode.GetDataComponents().Datapool().Proofs().AddProof(proof)
		}

		if proof.HeaderShardId != core.MetachainShardId {
			for _, shardNode := range s.consensusNodes[proof.HeaderShardId] {
				_ = shardNode.GetDataComponents().Datapool().Proofs().AddProof(proof)
			}
		}
	}
}

// createTestNodeWithKeys builds a node, optionally pinning the PEM file holding its managed
// validator keys. The validatorKeysPemOverride makes the node a single-key
// consensus participant (S5 phase B); an empty override keeps the multikey all-validators PEM.
func (s *simulator) createTestNodeWithKeys(
	outputConfigs configs.ArgsConfigsSimulator, args ArgsBaseChainSimulator, shardIDStr string, monitor factory.HeartbeatV2Monitor,
	validatorKeysPemOverride string,
) (process.NodeHandler, error) {
	nodeConfigs := cloneConstructionConfigs(outputConfigs.Configs)
	argsTestOnlyProcessorNode := components.ArgsTestOnlyProcessingNode{
		Configs:                      nodeConfigs,
		ChanStopNodeProcess:          s.chanStopNodeProcess,
		SyncedBroadcastNetwork:       s.syncedBroadcastNetwork,
		NumShards:                    s.numOfShards,
		GasScheduleFilename:          outputConfigs.GasScheduleFilename,
		ShardIDStr:                   shardIDStr,
		APIInterface:                 args.ApiInterface,
		BypassTxSignatureCheck:       args.BypassTxSignatureCheck,
		BypassBlockSignatureCheck:    args.BypassBlockSignatureCheck,
		InitialRound:                 args.InitialRound,
		InitialNonce:                 args.InitialNonce,
		MinNodesPerShard:             args.MinNodesPerShard,
		ConsensusGroupSize:           args.ConsensusGroupSize,
		MinNodesMeta:                 args.MetaChainMinNodes,
		MetaChainConsensusGroupSize:  args.MetaChainConsensusGroupSize,
		RoundDurationInMillis:        args.RoundDurationInMillis,
		VmQueryDelayAfterStartInMs:   args.VmQueryDelayAfterStartInMs,
		GenesisTime:                  s.genesisTime,
		Monitor:                      monitor,
		BypassCreateBlockTimeCheck:   args.BypassCreateBlockTimeCheck,
		CreateBlockMaxTimePercent:    args.CreateBlockMaxTimePercent,
		EnableConsensus:              args.ConsensusMode.isEnabled(),
		EnableFastConsensusCrypto:    args.ConsensusMode.isFastCrypto(),
		ValidatorKeysPemFileOverride: validatorKeysPemOverride,
	}

	return components.NewTestOnlyProcessingNode(argsTestOnlyProcessorNode)
}

// cloneConstructionConfigs isolates the slices that node component constructors sort in place.
// Consensus mode constructs nodes concurrently, so sharing their backing arrays would race and
// could expose a partially sorted configuration to another node.
func cloneConstructionConfigs(source config.Configs) config.Configs {
	result := source

	generalConfig := *source.GeneralConfig
	generalConfig.GeneralSettings.ChainParametersByEpoch = append(
		[]config.ChainParametersByEpochConfig(nil),
		source.GeneralConfig.GeneralSettings.ChainParametersByEpoch...,
	)
	generalConfig.Versions.VersionsByEpochs = append(
		[]config.VersionByEpochs(nil),
		source.GeneralConfig.Versions.VersionsByEpochs...,
	)
	result.GeneralConfig = &generalConfig

	epochConfig := *source.EpochConfig
	epochConfig.GasSchedule.GasScheduleByEpochs = append(
		[]config.GasScheduleByEpochs(nil),
		source.EpochConfig.GasSchedule.GasScheduleByEpochs...,
	)
	result.EpochConfig = &epochConfig

	return result
}

// GenerateBlocks will generate the provided number of blocks
func (s *simulator) GenerateBlocks(numOfBlocks int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for idx := 0; idx < numOfBlocks; idx++ {
		err := s.produceRound()
		if err != nil {
			return err
		}
	}
	return nil
}

// produceRound advances one logical round: (1) the simulator's global round increments,
// (2) nodes increment their internal rounds, (3) blocks are created.
// Must be called under s.mutex.
func (s *simulator) produceRound() error {
	if s.enableConsensus {
		// consensus mode drives the chronology, which owns the round index via the manual sync
		// timer; the direct per-node round increment must not run (the round handler is the
		// production round.NewRound, not the manual increment one)
		return s.advanceConsensusOnAllNodes()
	}

	s.incrementRoundOnAllValidators()

	return s.allNodesCreateBlocks()
}

// consensusDriveNode is implemented by nodes built in consensus-path execution mode; the simulator
// type-asserts to it to start and await one chronology subround at a time.
type consensusDriveNode interface {
	AdvanceConsensusClock() error
	RearmConsensusRound() error
	StepConsensusSubround() error
	WaitConsensusSubround() error
	ConsensusDriveState() (subround int, generation uint64, err error)
}

// maxConsensusPassesPerRound bounds how many round-robin passes the drive makes over a shard's
// nodes before giving up on a round. Each pass steps every node one subround; the
// group-size-1 happy path needs a handful, larger groups need more for the proposal/signature/proof
// message ping-pong plus the one-step commit-detection lag. An exhausted budget is a legitimate
// empty round (no commit), not an error: a stuck proposer produces none.
const maxConsensusPassesPerRound = 64

const (
	deferredExecutionMutationBarrierTimeout = 5 * time.Second
	deferredExecutionMutationBarrierPoll    = 5 * time.Millisecond
)

// advanceConsensusOnAllNodes drives every chain through one consensus round. The groups normally
// run concurrently. At the Andromeda activation boundary, however, the first normal metachain
// block needs current-round shard headers as finality attestations. Drive the shard groups first in
// that one round and deliver their committed headers and proofs before starting the metachain
// group. This gives every metachain validator the same proof-backed view instead of making
// ProcessBlock race the shard groups' EndRound broadcasts.
// Must be called under s.mutex.
func (s *simulator) advanceConsensusOnAllNodes() error {
	s.syncConsensusEpochBeforeRound()

	metaStartTip, _ := s.maxTipAndSource(core.MetachainShardId)
	isActivationBoundary := s.hasCurrentShardFlagActivationHeader(common.AndromedaFlag)

	if isActivationBoundary {
		err := s.advanceActivationConsensusRound()
		if err != nil {
			return err
		}
	} else {
		err := s.advanceConsensusGroupsConcurrently(true)
		if err != nil {
			return err
		}
	}

	// A v2 leader commits before its equivalent-proof broadcast has necessarily reached another
	// physical node's proof pool. Copy the now-final artifacts first, then catch up lagging nodes;
	// doing these in the opposite order can leave one node a nonce behind until the next API call,
	// where it proposes a competing block from stale randomness.
	s.deliverCommittedHeadersAndProofs()
	for _, handler := range s.handlers {
		s.syncBehindNodes(handler.shardID)
	}
	s.deliverCommittedHeadersAndProofs()

	// The first normal metachain block after an epoch-start block can begin processing before the
	// same-round shard finality attestations exist. The activation ordering above should prevent
	// that race; retain the same-round retry as a safety net for a legitimately empty proposal.
	metaTip, _ := s.maxTipAndSource(core.MetachainShardId)
	if metaTip <= metaStartTip {
		err := s.retryShardConsensusRound(core.MetachainShardId)
		if err != nil {
			return err
		}
		s.deliverCommittedHeadersAndProofs()
		s.syncBehindNodes(core.MetachainShardId)
		s.deliverCommittedHeadersAndProofs()
	}

	return nil
}

// advanceActivationConsensusRound serializes only the dependency edge at the Andromeda boundary:
// shard consensus still runs concurrently across shards, then metachain consensus starts after all
// current-round shard finality artifacts have been delivered.
func (s *simulator) advanceActivationConsensusRound() error {
	err := s.advanceConsensusGroupsConcurrently(false)
	if err != nil {
		return err
	}

	s.deliverCommittedHeadersAndProofs()
	for _, handler := range s.handlers {
		if handler.shardID != core.MetachainShardId {
			s.syncBehindNodes(handler.shardID)
		}
	}
	s.deliverCommittedHeadersAndProofs()

	return s.runShardConsensusRound(core.MetachainShardId)
}

// advanceConsensusGroupsConcurrently drives either every group or only shard groups.
func (s *simulator) advanceConsensusGroupsConcurrently(includeMetachain bool) error {
	errChan := make(chan error, len(s.handlers))
	deliveryMutex := sync.Mutex{}
	numGroups := 0
	for _, handler := range s.handlers {
		shardID := handler.shardID
		if !includeMetachain && shardID == core.MetachainShardId {
			continue
		}
		numGroups++

		go func() {
			onTipAdvanced := func() {
				deliveryMutex.Lock()
				s.deliverCommittedHeadersAndProofs()
				deliveryMutex.Unlock()
			}

			errChan <- s.runShardConsensusRoundWithProgress(shardID, onTipAdvanced)
		}()
	}

	var firstErr error
	for idx := 0; idx < numGroups; idx++ {
		err := <-errChan
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// syncConsensusEpochBeforeRound makes every physical validator observe the committed metachain
// epoch before any chronology is stepped. Epoch notifications normally arrive asynchronously with
// metachain headers; in the in-process simulator that can leave one validator on consensus v1 while
// its peer already starts the same round on v2. The committed metachain header is the global epoch
// source, and CheckEpoch is idempotent when no transition is due.
func (s *simulator) syncConsensusEpochBeforeRound() {
	_, metaSource := s.maxTipAndSource(core.MetachainShardId)
	if check.IfNilReflect(metaSource) {
		return
	}

	metaHeader := metaSource.GetChainHandler().GetCurrentBlockHeader()
	if check.IfNil(metaHeader) {
		return
	}

	for shardID := range s.consensusNodes {
		s.syncedBroadcastNetwork.NotifyHeader(shardID, metaHeader)
	}
}

// hasCurrentShardFlagActivationHeader identifies the one transition round in which the metachain
// validates shard activation headers using legacy finality rules. Those rules can make a minority
// of metachain validators wait for current-round shard attestations after the PBFT quorum has
// already processed the proposal.
func (s *simulator) hasCurrentShardFlagActivationHeader(flag core.EnableEpochFlag) bool {
	for shardID := uint32(0); shardID < s.numOfShards; shardID++ {
		_, source := s.maxTipAndSource(shardID)
		if check.IfNilReflect(source) {
			continue
		}

		header := source.GetChainHandler().GetCurrentBlockHeader()
		if check.IfNil(header) {
			continue
		}

		if common.IsEpochChangeBlockForFlagActivation(
			header,
			source.GetCoreComponents().EnableEpochsHandler(),
			flag,
		) {
			return true
		}
	}

	return false
}

// deliverCommittedHeadersAndProofs feeds every node's freshly committed header into every other node's
// header pool. Each leader broadcasts its committed header from its EndRound subround, which runs in the
// chronology goroutine, so when nodes are stepped in sequence within a round a broadcast can land after
// the recipient has already proposed and be missed. That matters most for notarization, which proceeds
// only as an unbroken chain from the last notarized header: a single missed shard header blocks ALL
// further metachain notarization of that shard (its self-notarized-by-meta nonce freezes while it keeps
// producing), the gap crosses MaxShardNoncesBehind, the "shard is stuck" protection trips, and
// cross-shard transfers stall until the wall-clock requester back-fills the gap hundreds of rounds later
// — a source of unreliable cross-shard latency. The same race in the reverse direction
// keeps a shard from learning it was notarized. Delivering committed headers both ways here removes the
// race from the manual drive.
//
// Every header is a real committed block, validated and notarized normally on receipt, so this only
// guarantees delivery; it does not bypass block validation.
//
// The block's v2/Andromeda equivalent proof travels the same way: the metachain notarizes a shard block
// (and a shard notarizes a metablock) only once it holds that block's equivalent proof, so a missed proof
// freezes notarization exactly like a missed header. Deliver both the current proof, when already
// finalized, and the previous header's proof. Must be called under s.mutex.
func (s *simulator) deliverCommittedHeadersAndProofs() {
	type committedBlock struct {
		hash         []byte
		shardID      uint32
		header       data.HeaderHandler
		currentProof data.HeaderProofHandler
		prevHash     []byte
		prevProof    data.HeaderProofHandler
	}

	committed := make([]committedBlock, 0, len(s.consensusNodes))
	for shardID := range s.consensusNodes {
		_, src := s.maxTipAndSource(shardID)
		if check.IfNilReflect(src) {
			continue
		}

		chain := src.GetChainHandler()
		header := chain.GetCurrentBlockHeader()
		if check.IfNil(header) {
			continue
		}
		hash := chain.GetCurrentBlockHeaderHash()
		proofsPool := src.GetDataComponents().Datapool().Proofs()
		currentProof, _ := proofsPool.GetProof(header.GetShardID(), hash)
		prevHash := header.GetPrevHash()
		prevProof, _ := proofsPool.GetProof(header.GetShardID(), prevHash)

		// Nodes commit the same header at slightly different instants. The first max-tip node is
		// therefore not necessarily the one whose proof pool already contains the equivalent proof.
		// Search the whole shard group before delivering; otherwise the one tip-advance callback can
		// copy the header without its proof and leave some metachain validators waiting until their
		// BLOCK deadline even though another shard node already holds the proof.
		for _, candidate := range s.consensusNodes[shardID] {
			candidateProofs := candidate.GetDataComponents().Datapool().Proofs()
			if check.IfNil(currentProof) {
				currentProof, _ = candidateProofs.GetProof(header.GetShardID(), hash)
			}
			if check.IfNil(prevProof) {
				prevProof, _ = candidateProofs.GetProof(header.GetShardID(), prevHash)
			}
			if !check.IfNil(currentProof) && !check.IfNil(prevProof) {
				break
			}
		}
		committed = append(committed, committedBlock{
			hash:         hash,
			shardID:      header.GetShardID(),
			header:       header,
			currentProof: currentProof,
			prevHash:     prevHash,
			prevProof:    prevProof,
		})
	}

	for _, nodes := range s.consensusNodes {
		for _, dst := range nodes {
			headersPool := dst.GetDataComponents().Datapool().Headers()
			proofsPool := dst.GetDataComponents().Datapool().Proofs()
			for _, c := range committed {
				if _, err := headersPool.GetHeaderByHash(c.hash); err != nil {
					headersPool.AddHeader(c.hash, c.header)
				}
				if !check.IfNil(c.currentProof) {
					deliveryKey := dst.GetNetworkComponents().NetworkMessenger().ID().Pretty() +
						":" + string(c.hash)
					_, delivered := s.deliveredConsensusProofs[deliveryKey]
					if !delivered {
						_ = proofsPool.UpsertProof(c.currentProof)
						s.deliveredConsensusProofs[deliveryKey] = struct{}{}
					}
				}
				if !check.IfNil(c.prevProof) {
					if !proofsPool.HasProof(c.shardID, c.prevHash) {
						proofsPool.AddProof(c.prevProof)
					}
				}
			}
		}
	}
}

// runShardConsensusRound drives all of a shard's consensus nodes through one round. It advances
// every node's clock, then starts one subround on every validator before waiting for any of them.
// This preserves concurrent proposal/signature exchange without allowing one validator's chronology
// to overtake another.
func (s *simulator) runShardConsensusRound(shardID uint32) error {
	return s.runShardConsensusRoundWithProgress(shardID, nil)
}

// runShardConsensusRoundWithProgress is runShardConsensusRound with a callback invoked whenever
// this group's committed tip advances. The simulator uses it to deliver shard finality artifacts
// while the metachain group is still processing the same round.
func (s *simulator) runShardConsensusRoundWithProgress(
	shardID uint32,
	onTipAdvanced func(),
) error {
	return s.driveShardConsensusRound(shardID, onTipAdvanced, true)
}

// retryShardConsensusRound repeats consensus at the current manual-clock round. It is used only
// after all shard groups have finished and delivered the finality artifacts that a metachain
// attempt was missing.
func (s *simulator) retryShardConsensusRound(shardID uint32) error {
	return s.driveShardConsensusRound(shardID, nil, false)
}

func (s *simulator) driveShardConsensusRound(
	shardID uint32,
	onTipAdvanced func(),
	advanceClock bool,
) error {
	driveStarted := time.Now()
	nodes := s.consensusNodes[shardID]

	drivers := make([]consensusDriveNode, 0, len(nodes))
	driverNodes := make([]process.NodeHandler, 0, len(nodes))
	for _, n := range nodes {
		driver, ok := n.(consensusDriveNode)
		if !ok {
			return errConsensusNodeNotDriveable
		}
		if advanceClock {
			err := driver.AdvanceConsensusClock()
			if err != nil {
				return err
			}
		} else {
			// A failed subround leaves the chronology before START_ROUND. Since the manual clock
			// intentionally stays on the same round for a retry, the chronology would otherwise see
			// no round-index change and every retry tick would be a no-op. Revert the round handler
			// once so the next tick updates back to the current round and initializes its subrounds,
			// exactly as the production chronology does when it (re)starts.
			err := driver.RearmConsensusRound()
			if err != nil {
				return err
			}
		}

		drivers = append(drivers, driver)
		driverNodes = append(driverNodes, n)
	}

	// Keep stepping until this logical round commits one block. Physical peers need not all finish
	// the commit before the drive returns: the post-round delivery/catch-up phase applies the real
	// committed header and proof to laggards. Continuing to step after the first commit would let an
	// already-advanced validator propose nonce+1 at this SAME manual round while a peer is still
	// completing END_ROUND for the committed block.
	startTip := maxTipOf(driverNodes)
	lastReportedTip := startTip
	generations, err := consensusDriveGenerations(drivers)
	if err != nil {
		return err
	}
	catchingUpAfterRestart := false
	for pass := 0; pass < maxConsensusPassesPerRound; pass++ {
		passStarted := time.Now()
		driversToStep := drivers
		if catchingUpAfterRestart {
			var caughtUp bool
			driversToStep, caughtUp, err = laggingConsensusDrivers(drivers)
			if err != nil {
				return err
			}
			if caughtUp {
				catchingUpAfterRestart = false
				driversToStep = drivers
			}
		}

		for _, driver := range driversToStep {
			err := driver.StepConsensusSubround()
			if err != nil {
				return err
			}
		}

		waitResultChan := make(chan error, len(driversToStep))
		for _, driver := range driversToStep {
			driver := driver
			go func() {
				waitResultChan <- driver.WaitConsensusSubround()
			}()
		}

		var firstWaitErr error
		for range driversToStep {
			waitErr := <-waitResultChan
			if waitErr != nil && firstWaitErr == nil {
				firstWaitErr = waitErr
			}

			restarted, err := updateConsensusDriveGenerations(drivers, generations)
			if err != nil {
				return err
			}
			if restarted {
				catchingUpAfterRestart = true
				s.syncConsensusEpochAfterRestart(shardID)
			}

			// A consensus-group member can commit while a reshuffled-out physical node is still
			// finishing the same subround. Report that tip immediately: waiting for every node here
			// postpones the attesting header until after the metachain BLOCK deadline.
			tip := maxTipOf(driverNodes)
			if tip > lastReportedTip && onTipAdvanced != nil {
				onTipAdvanced()
				lastReportedTip = tip
			}
		}
		if firstWaitErr != nil {
			return firstWaitErr
		}

		restarted, err := updateConsensusDriveGenerations(drivers, generations)
		if err != nil {
			return err
		}
		catchingUpAfterRestart = catchingUpAfterRestart || restarted

		passDuration := time.Since(passStarted)
		if passDuration > 50*time.Millisecond {
			subrounds := make([]int, len(drivers))
			for idx, driver := range drivers {
				subrounds[idx], _, _ = driver.ConsensusDriveState()
			}
			log.Trace("slow consensus drive pass",
				"shard", shardID,
				"pass", pass,
				"duration", passDuration,
				"subrounds", subrounds,
				"start tip", startTip,
				"current tip", maxTipOf(driverNodes),
			)
		}

		if maxTipOf(driverNodes) > startTip {
			// The first tip observation can happen while another validator is still finishing
			// EndRound. In v2 that validator may be the one assembling the equivalent proof, so
			// deliver once more after every stepped validator has acknowledged the subround.
			// Metachain ProcessBlock waits on these shard finality artifacts.
			if onTipAdvanced != nil {
				onTipAdvanced()
			}
			s.logConsensusLeader(shardID)
			break
		}
	}

	log.Trace("consensus shard drive completed",
		"shard", shardID,
		"advance clock", advanceClock,
		"start tip", startTip,
		"end tip", maxTipOf(driverNodes),
		"duration", time.Since(driveStarted),
	)

	return nil
}

// syncConsensusEpochAfterRestart propagates an epoch that one real consensus participant has
// already discovered while creating the activation header. The v1 leader's epoch notifier switches
// its chronology to v2 before the old v1 BLOCK job returns; peers still waiting in that obsolete
// BLOCK subround otherwise consume their full wall-clock deadline before they can observe v2.
//
// Notify only when the shard's physical nodes actually disagree, and use the highest epoch already
// reported by a node as the source. NotifyHeader reaches the same per-node generic epoch notifier
// used by a received header. Closing the old chronology generation cancels its in-flight context,
// so every peer can restart on v2 without waiting for a subround that can no longer succeed.
func (s *simulator) syncConsensusEpochAfterRestart(shardID uint32) {
	nodes := s.consensusNodes[shardID]
	if len(nodes) == 0 {
		return
	}

	minEpoch := nodes[0].GetCoreComponents().EpochNotifier().CurrentEpoch()
	maxEpoch := minEpoch
	var maxEpochSource process.NodeHandler = nodes[0]
	for _, node := range nodes[1:] {
		epoch := node.GetCoreComponents().EpochNotifier().CurrentEpoch()
		if epoch < minEpoch {
			minEpoch = epoch
		}
		if epoch > maxEpoch {
			maxEpoch = epoch
			maxEpochSource = node
		}
	}
	if minEpoch == maxEpoch {
		return
	}

	roundHandler := maxEpochSource.GetCoreComponents().RoundHandler()
	epochHeader := &block.MetaBlock{
		Epoch:     maxEpoch,
		Round:     uint64(roundHandler.Index()),
		TimeStamp: uint64(roundHandler.TimeStamp().Unix()),
	}
	s.syncedBroadcastNetwork.NotifyHeader(shardID, epochHeader)
}

// logConsensusLeader preserves the simulator's existing "Leader in current block" diagnostic in
// consensus mode. Direct mode emits this line while it computes the synthetic header; consensus
// computes the same group inside SPoS, so emit the equivalent self-contained record only after a
// real block was committed. Testing-suite fee scenarios use it to resolve the reward recipient.
func (s *simulator) logConsensusLeader(shardID uint32) {
	_, source := s.maxTipAndSource(shardID)
	if check.IfNilReflect(source) {
		return
	}

	header := source.GetChainHandler().GetCurrentBlockHeader()
	if check.IfNil(header) {
		return
	}

	leader, _, err := source.GetProcessComponents().NodesCoordinator().ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch(),
	)
	if err != nil {
		log.Debug("could not compute leader for committed consensus block",
			"shardID", header.GetShardID(),
			"round", header.GetRound(),
			"error", err)
		return
	}

	log.Debug("Leader in current block",
		"shardID", header.GetShardID(),
		"round", header.GetRound(),
		"leader", leader.PubKey())
}

func consensusDriveGenerations(drivers []consensusDriveNode) ([]uint64, error) {
	generations := make([]uint64, len(drivers))
	for idx, driver := range drivers {
		_, generation, err := driver.ConsensusDriveState()
		if err != nil {
			return nil, err
		}
		generations[idx] = generation
	}

	return generations, nil
}

func updateConsensusDriveGenerations(drivers []consensusDriveNode, generations []uint64) (bool, error) {
	restarted := false
	for idx, driver := range drivers {
		_, generation, err := driver.ConsensusDriveState()
		if err != nil {
			return false, err
		}
		if generation != generations[idx] {
			restarted = true
			generations[idx] = generation
		}
	}

	return restarted, nil
}

// laggingConsensusDrivers selects only chronologies behind the most advanced chronology. This is
// used briefly after a consensus-version restart: stepping every physical node would preserve the
// subround offset and make the advanced node enter END_ROUND before the restarted node can sign.
func laggingConsensusDrivers(drivers []consensusDriveNode) ([]consensusDriveNode, bool, error) {
	if len(drivers) == 0 {
		return nil, true, nil
	}

	subrounds := make([]int, len(drivers))
	mostAdvanced := 0
	for idx, driver := range drivers {
		subround, _, err := driver.ConsensusDriveState()
		if err != nil {
			return nil, false, err
		}
		subrounds[idx] = subround
		if idx == 0 || subround > mostAdvanced {
			mostAdvanced = subround
		}
	}

	lagging := make([]consensusDriveNode, 0, len(drivers))
	for idx, subround := range subrounds {
		if subround < mostAdvanced {
			lagging = append(lagging, drivers[idx])
		}
	}

	return lagging, len(lagging) == 0, nil
}

// GenerateBlocksUntilEpochIsReached will generate blocks until the epoch is reached
func (s *simulator) GenerateBlocksUntilEpochIsReached(targetEpoch int32) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	maxNumberOfRounds := 10000
	for idx := 0; idx < maxNumberOfRounds; idx++ {
		err := s.produceRound()
		if err != nil {
			return err
		}

		epochReachedOnAllNodes, errEpoch := s.isTargetEpochReached(targetEpoch)
		if errEpoch != nil {
			return errEpoch
		}

		if epochReachedOnAllNodes {
			return nil
		}
	}
	return fmt.Errorf("exceeded rounds to generate blocks")
}

// ForceResetValidatorStatisticsCache will force the reset of the cache used for the validators statistics endpoint
func (s *simulator) ForceResetValidatorStatisticsCache() error {
	metachainNode := s.GetNodeHandler(core.MetachainShardId)
	if check.IfNil(metachainNode) {
		return errNilMetachainNode
	}

	return metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
}

func (s *simulator) isTargetEpochReached(targetEpoch int32) (bool, error) {
	metachainNode := s.nodes[core.MetachainShardId]
	metachainEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()

	for shardID, n := range s.nodes {
		if shardID != core.MetachainShardId {
			if int32(n.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()) < int32(metachainEpoch-1) {
				return false, fmt.Errorf("shard %d is with at least 2 epochs behind metachain shard node epoch %d, metachain node epoch %d",
					shardID, n.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch(), metachainEpoch)
			}
		}

		if int32(n.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()) < targetEpoch {
			return false, nil
		}
	}

	if s.enableConsensus {
		// Epoch notifiers run while an epoch-start proposal is broadcast, before every shard has
		// committed its first block in that epoch. Returning based on the notifier alone races API
		// callers against that commit: reward miniblocks can still be unapplied, so an immediate
		// balance query nondeterministically observes the previous epoch. In consensus mode, require
		// the primary of every chain to have committed an actual target-epoch block.
		for _, n := range s.nodes {
			header := n.GetChainHandler().GetCurrentBlockHeader()
			if check.IfNil(header) || int32(header.GetEpoch()) < targetEpoch {
				return false, nil
			}
		}
	}

	return true, nil
}

func (s *simulator) incrementRoundOnAllValidators() {
	for _, node := range s.handlers {
		node.handler.IncrementRound()
	}
}

// ForceChangeOfEpoch advances the chain by one epoch. In direct mode it arms each node's epoch-start
// trigger and generates blocks until the next epoch commits. In consensus-path execution mode the epoch
// boundary is owned by the chronology and can only be reached by producing rounds — the trigger cannot be
// forced mid-epoch without desyncing the consensus group — so it drives rounds to the next natural epoch
// boundary instead. Either way the caller ends one epoch further along (consensus mode just plays out the
// remaining rounds of the current epoch).
func (s *simulator) ForceChangeOfEpoch() error {
	s.mutex.Lock()

	if s.enableConsensus {
		currentEpoch := s.nodes[core.MetachainShardId].GetProcessComponents().EpochStartTrigger().Epoch()
		s.mutex.Unlock()
		// GenerateBlocksUntilEpochIsReached takes the mutex itself
		return s.GenerateBlocksUntilEpochIsReached(int32(currentEpoch + 1))
	}

	log.Info("force change of epoch")
	for shardID, node := range s.nodes {
		err := node.ForceChangeOfEpoch()
		if err != nil {
			s.mutex.Unlock()
			return fmt.Errorf("force change of epoch shardID-%d: error-%w", shardID, err)
		}
	}

	epoch := s.nodes[core.MetachainShardId].GetProcessComponents().EpochStartTrigger().Epoch()
	s.mutex.Unlock()

	err := s.GenerateBlocksUntilEpochIsReached(int32(epoch + 1))
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.produceRound()
}

func (s *simulator) allNodesCreateBlocks() error {
	headers := make(map[uint32]*dtos.BroadcastData, len(s.handlers))
	for _, node := range s.handlers {
		// TODO MX-15150 remove this when we remove all goroutines
		time.Sleep(2 * time.Millisecond)

		pair, err := node.handler.CreateNewBlock()
		if err != nil {
			return err
		}
		if pair == nil {
			continue
		}

		headers[pair.Header.GetShardID()] = pair
	}

	for shardID, pair := range headers {
		messenger := s.nodes[shardID].GetBroadcastMessenger()

		err := messenger.BroadcastHeader(pair.Header, pair.LeaderKey)
		if err != nil {
			return err
		}

		err = messenger.BroadcastMiniBlocks(pair.MiniBlocksBytes, pair.LeaderKey)
		if err != nil {
			return err
		}

		err = messenger.BroadcastTransactions(pair.TransactionsBytes, pair.LeaderKey)
		if err != nil {
			return err
		}

		if !check.IfNil(pair.Proof) {
			time.Sleep(time.Millisecond * 5) // small delay to ensure proof is not dropped as being received before header
			err = s.nodes[shardID].GetBroadcastMessenger().BroadcastEquivalentProof(pair.Proof, pair.LeaderKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetNodeHandler returns the node handler from the provided shardID
func (s *simulator) GetNodeHandler(shardID uint32) process.NodeHandler {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.nodes[shardID]
}

// GetRestAPIInterfaces will return a map with the rest api interfaces for every node
func (s *simulator) GetRestAPIInterfaces() map[uint32]string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	resMap := make(map[uint32]string)
	for shardID, node := range s.nodes {
		resMap[shardID] = node.GetFacadeHandler().RestApiInterface()
	}

	return resMap
}

// GetInitialWalletKeys will return the initial wallet keys
func (s *simulator) GetInitialWalletKeys() *dtos.InitialWalletKeys {
	return s.initialWalletKeys
}

// AddValidatorKeys will add the provided validators private keys in the keys handler on all nodes
func (s *simulator) AddValidatorKeys(validatorsPrivateKeys [][]byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, node := range s.nodes {
		for _, privateKey := range validatorsPrivateKeys {
			err := s.addManagedValidatorKey(node, privateKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GenerateAndMintWalletAddress will generate an address in the provided shard and will mint that address with the provided value
// if the target shard ID value does not correspond to a node handled by the chain simulator, the address will be generated in a random shard ID
func (s *simulator) GenerateAndMintWalletAddress(targetShardID uint32, value *big.Int) (dtos.WalletAddress, error) {
	wallet := s.GenerateAddressInShard(targetShardID)

	err := s.SetStateMultiple([]*dtos.AddressState{
		{
			Address: wallet.Bech32,
			Balance: value.String(),
		},
	})

	return wallet, err
}

// GenerateAddressInShard will generate a wallet address based on the provided shard
func (s *simulator) GenerateAddressInShard(providedShardID uint32) dtos.WalletAddress {
	converter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
	nodeHandler := s.GetNodeHandler(providedShardID)
	if check.IfNil(nodeHandler) {
		return generateWalletAddress(converter)
	}

	for {
		buff := generateAddress(converter.Len())
		if nodeHandler.GetShardCoordinator().ComputeId(buff) == providedShardID {
			return generateWalletAddressFromBuffer(converter, buff)
		}
	}
}

func generateWalletAddress(converter core.PubkeyConverter) dtos.WalletAddress {
	buff := generateAddress(converter.Len())
	return generateWalletAddressFromBuffer(converter, buff)
}

func generateWalletAddressFromBuffer(converter core.PubkeyConverter, buff []byte) dtos.WalletAddress {
	return dtos.WalletAddress{
		Bech32: converter.SilentEncode(buff, log),
		Bytes:  buff,
	}
}

func generateAddress(len int) []byte {
	buff := make([]byte, len)
	_, _ = rand.Read(buff)

	return buff
}

func (s *simulator) setValidatorKeysForNode(node process.NodeHandler, validatorsPrivateKeys [][]byte) error {
	for idx, privateKey := range validatorsPrivateKeys {
		err := s.addManagedValidatorKey(node, privateKey)
		if err != nil {
			return fmt.Errorf("cannot add private key for shard=%d, index=%d, error=%s", node.GetShardCoordinator().SelfId(), idx, err.Error())
		}
	}

	return nil
}

func (s *simulator) addManagedValidatorKey(node process.NodeHandler, privateKeyBytes []byte) error {
	err := node.GetCryptoComponents().ManagedPeersHolder().AddManagedPeer(privateKeyBytes)
	if err != nil {
		return err
	}

	keyGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	privateKey, err := keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	if err != nil {
		return err
	}
	publicKeyBytes, err := privateKey.GeneratePublic().ToByteArray()
	if err != nil {
		return err
	}

	_, virtualPID, err := node.GetCryptoComponents().KeysHandler().GetP2PIdentity(publicKeyBytes)
	if err != nil {
		return err
	}
	physicalPID := node.GetNetworkComponents().NetworkMessenger().ID()

	return s.syncedBroadcastNetwork.RegisterPeerAlias(virtualPID, physicalPID)
}

// GetValidatorPrivateKeys will return the initial validators private keys
func (s *simulator) GetValidatorPrivateKeys() []crypto.PrivateKey {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.validatorsPrivateKeys
}

// SetKeyValueForAddress will set the provided state for a given address
func (s *simulator) SetKeyValueForAddress(address string, keyValueMap map[string]string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.waitForDeferredExecutionBeforeStateMutation()
	if err != nil {
		return err
	}

	addressConverter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
	addressBytes, err := addressConverter.Decode(address)
	if err != nil {
		return err
	}

	if bytes.Equal(addressBytes, core.SystemAccountAddress) {
		return s.setKeyValueSystemAccount(keyValueMap)
	}

	shardID := sharding.ComputeShardID(addressBytes, s.numOfShards)
	shardNodes, ok := s.consensusNodes[shardID]
	if !ok {
		return fmt.Errorf("cannot find a test node for the computed shard id, computed shard id: %d", shardID)
	}

	for _, node := range shardNodes {
		err = node.SetKeyValueForAddress(addressBytes, keyValueMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *simulator) setKeyValueSystemAccount(keyValueMap map[string]string) error {
	for shard, nodes := range s.consensusNodes {
		for _, node := range nodes {
			err := node.SetKeyValueForAddress(core.SystemAccountAddress, keyValueMap)
			if err != nil {
				return fmt.Errorf("%w for shard %d", err, shard)
			}
		}
	}

	return nil
}

// SetStateMultiple will set state for multiple addresses
func (s *simulator) SetStateMultiple(stateSlice []*dtos.AddressState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.waitForDeferredExecutionBeforeStateMutation()
	if err != nil {
		return err
	}

	addressConverter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
	for _, stateValue := range stateSlice {
		addressBytes, err := addressConverter.Decode(stateValue.Address)
		if err != nil {
			return err
		}

		if bytes.Equal(addressBytes, core.SystemAccountAddress) {
			err = s.setStateSystemAccount(stateValue)
		} else {
			// every consensus node of the shard must hold the state, not just the query primary:
			// the elected leader (any node) reads from its own accounts when validating txs
			shardID := sharding.ComputeShardID(addressBytes, s.numOfShards)
			for _, node := range s.consensusNodes[shardID] {
				err = node.SetStateForAddress(addressBytes, stateValue)
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveAccounts will try to remove all accounts data for the addresses provided
func (s *simulator) RemoveAccounts(addresses []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.waitForDeferredExecutionBeforeStateMutation()
	if err != nil {
		return err
	}

	addressConverter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
	for _, address := range addresses {
		addressBytes, err := addressConverter.Decode(address)
		if err != nil {
			return err
		}

		if bytes.Equal(addressBytes, core.SystemAccountAddress) {
			err = s.removeAllSystemAccounts()
		} else {
			shardID := sharding.ComputeShardID(addressBytes, s.numOfShards)
			for _, node := range s.consensusNodes[shardID] {
				err = node.RemoveAccount(addressBytes)
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// waitForDeferredExecutionBeforeStateMutation prevents the HeaderV3 executor from committing an
// already queued block over state injected through the simulator API. Consensus commits and
// deferred execution are deliberately decoupled under Supernova, so a successful GenerateBlocks
// call does not by itself guarantee that it is safe to alter the accounts trie. Direct mode and
// pre-Supernova blocks do not need this barrier.
func (s *simulator) waitForDeferredExecutionBeforeStateMutation() error {
	if !s.enableConsensus {
		return nil
	}

	timer := time.NewTimer(deferredExecutionMutationBarrierTimeout)
	defer timer.Stop()

	ticker := time.NewTicker(deferredExecutionMutationBarrierPoll)
	defer ticker.Stop()

	for {
		var laggingShard uint32
		var laggingNode int
		var committedNonce uint64
		var executedNonce uint64
		allCaughtUp := true

		for shardID, nodes := range s.consensusNodes {
			for nodeIndex, node := range nodes {
				currentHeader := node.GetChainHandler().GetCurrentBlockHeader()
				if check.IfNil(currentHeader) || !currentHeader.IsHeaderV3() {
					continue
				}

				lastExecutedNonce, _, _ := node.GetChainHandler().GetLastExecutedBlockInfo()
				if lastExecutedNonce >= currentHeader.GetNonce() {
					continue
				}

				allCaughtUp = false
				laggingShard = shardID
				laggingNode = nodeIndex
				committedNonce = currentHeader.GetNonce()
				executedNonce = lastExecutedNonce
				break
			}
			if !allCaughtUp {
				break
			}
		}

		if allCaughtUp {
			return nil
		}

		select {
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf(
				"timed out waiting for deferred execution before state mutation: shard %d node %d committed nonce %d, executed nonce %d",
				laggingShard,
				laggingNode,
				committedNonce,
				executedNonce,
			)
		}
	}
}

// SendTxAndGenerateBlockTilTxIsExecuted will send the provided transaction and generate block until the transaction is executed
func (s *simulator) SendTxAndGenerateBlockTilTxIsExecuted(txToSend *transaction.Transaction, maxNumOfBlocksToGenerateWhenExecutingTx int) (*transaction.ApiTransactionResult, error) {
	result, err := s.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txToSend}, maxNumOfBlocksToGenerateWhenExecutingTx)
	if err != nil {
		return nil, err
	}

	return result[0], nil
}

// SendTxsAndGenerateBlocksTilAreExecuted will send the provided transactions and generate block until all transactions are executed
func (s *simulator) SendTxsAndGenerateBlocksTilAreExecuted(txsToSend []*transaction.Transaction, maxNumOfBlocksToGenerateWhenExecutingTx int) ([]*transaction.ApiTransactionResult, error) {
	if len(txsToSend) == 0 {
		return nil, chainSimulatorErrors.ErrEmptySliceOfTxs
	}
	if maxNumOfBlocksToGenerateWhenExecutingTx == 0 {
		return nil, chainSimulatorErrors.ErrInvalidMaxNumOfBlocks
	}

	transactionStatus := make([]*transactionWithResult, 0, len(txsToSend))
	for idx, tx := range txsToSend {
		if tx == nil {
			return nil, fmt.Errorf("%w on position %d", chainSimulatorErrors.ErrNilTransaction, idx)
		}

		txHashHex, err := s.sendTx(tx)
		if err != nil {
			return nil, err
		}

		transactionStatus = append(transactionStatus, &transactionWithResult{
			hexHash: txHashHex,
			tx:      tx,
		})
	}

	time.Sleep(delaySendTxs)

	for count := 0; count < maxNumOfBlocksToGenerateWhenExecutingTx; count++ {
		err := s.GenerateBlocks(1)
		if err != nil {
			return nil, err
		}

		txsAreExecuted := s.computeTransactionsStatus(transactionStatus)
		if txsAreExecuted {
			return getApiTransactionsFromResult(transactionStatus), nil
		}
	}

	return nil, errors.New("something went wrong. Transaction(s) is/are still in pending")
}

func (s *simulator) computeTransactionsStatus(txsWithResult []*transactionWithResult) bool {
	allAreExecuted := true
	contractDeploySCAddress := make([]byte, s.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Len())
	for _, resultTx := range txsWithResult {
		if resultTx.result != nil {
			continue
		}

		sentTx := resultTx.tx
		destinationShardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(sentTx.RcvAddr)
		if bytes.Equal(sentTx.RcvAddr, contractDeploySCAddress) {
			destinationShardID = s.GetNodeHandler(0).GetShardCoordinator().ComputeId(sentTx.SndAddr)
		}

		result, errGet := s.GetNodeHandler(destinationShardID).GetFacadeHandler().GetTransaction(resultTx.hexHash, true)
		if errGet == nil && result.Status != transaction.TxStatusPending {
			log.Trace("############## transaction was executed ##############", "txHash", resultTx.hexHash)
			resultTx.result = result
			continue
		}

		allAreExecuted = false
	}

	return allAreExecuted
}

func getApiTransactionsFromResult(txWithResult []*transactionWithResult) []*transaction.ApiTransactionResult {
	result := make([]*transaction.ApiTransactionResult, 0, len(txWithResult))
	for _, tx := range txWithResult {
		result = append(result, tx.result)
	}

	return result
}

func (s *simulator) sendTx(tx *transaction.Transaction) (string, error) {
	shardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(tx.SndAddr)
	err := s.GetNodeHandler(shardID).GetFacadeHandler().ValidateTransaction(tx)
	if err != nil {
		return "", err
	}

	node := s.GetNodeHandler(shardID)
	txHash, err := core.CalculateHash(node.GetCoreComponents().InternalMarshalizer(), node.GetCoreComponents().Hasher(), tx)
	if err != nil {
		return "", err
	}

	txHashHex := hex.EncodeToString(txHash)
	_, err = node.GetFacadeHandler().SendBulkTransactions([]*transaction.Transaction{tx})
	if err != nil {
		return "", err
	}

	for {
		recoveredTx, _ := node.GetFacadeHandler().GetTransaction(txHashHex, false)
		if recoveredTx != nil {
			log.Trace("############## send transaction ##############", "txHash", txHashHex)
			return txHashHex, nil
		}

		time.Sleep(delaySendTxs)
	}
}

func (s *simulator) setStateSystemAccount(state *dtos.AddressState) error {
	for shard, nodes := range s.consensusNodes {
		for _, node := range nodes {
			err := node.SetStateForAddress(core.SystemAccountAddress, state)
			if err != nil {
				return fmt.Errorf("%w for shard %d", err, shard)
			}
		}
	}

	return nil
}

func (s *simulator) removeAllSystemAccounts() error {
	for shard, nodes := range s.consensusNodes {
		for _, node := range nodes {
			err := node.RemoveAccount(core.SystemAccountAddress)
			if err != nil {
				return fmt.Errorf("%w for shard %d", err, shard)
			}
		}
	}

	return nil
}

// initializeMetachainGenesisState applies the metachain genesis system-SC processing the
// simulator performs at construction, committing the resulting accounts state; it must run
// on every freshly built metachain node whose chain is at genesis
func initializeMetachainGenesisState(node process.NodeHandler) error {
	currentRootHash, err := node.GetProcessComponents().ValidatorsStatistics().RootHash()
	if err != nil {
		return err
	}

	allValidatorsInfo, err := node.GetProcessComponents().ValidatorsStatistics().GetValidatorInfoForRootHash(currentRootHash)
	if err != nil {
		return err
	}

	err = node.GetProcessComponents().EpochSystemSCProcessor().ProcessSystemSmartContract(
		allValidatorsInfo,
		node.GetDataComponents().Blockchain().GetGenesisHeader(),
	)
	if err != nil {
		return err
	}

	_, err = node.GetStateComponents().AccountsAdapter().Commit()

	return err
}

// GetAccount will fetch the account of the provided address
func (s *simulator) GetAccount(address dtos.WalletAddress) (api.AccountResponse, error) {
	destinationShardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	account, _, err := s.GetNodeHandler(destinationShardID).GetFacadeHandler().GetAccount(address.Bech32, api.AccountQueryOptions{})
	return account, err
}

// Close will stop and close the simulator
func (s *simulator) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errorStrings []string
	// close every consensus node, not only the per-shard query primary: a multi-validator group's
	// extra nodes live only in consensusNodes and own chronology goroutines, network registrations,
	// state and storage close handlers that would otherwise leak
	for _, nodes := range s.consensusNodes {
		for _, n := range nodes {
			err := n.Close()
			if err != nil {
				errorStrings = append(errorStrings, err.Error())
			}
		}
	}

	if len(errorStrings) != 0 {
		log.Error("error closing chain simulator", "error", components.AggregateErrors(errorStrings, components.ErrClose))
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simulator) IsInterfaceNil() bool {
	return s == nil
}

// GenerateBlsPrivateKeys will generate bls keys
func GenerateBlsPrivateKeys(numOfKeys int) ([][]byte, []string, error) {
	blockSigningGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	secretKeysBytes := make([][]byte, 0, numOfKeys)
	blsKeysHex := make([]string, 0, numOfKeys)
	for idx := 0; idx < numOfKeys; idx++ {
		secretKey, publicKey := blockSigningGenerator.GeneratePair()

		secretKeyBytes, err := secretKey.ToByteArray()
		if err != nil {
			return nil, nil, err
		}

		secretKeysBytes = append(secretKeysBytes, secretKeyBytes)

		publicKeyBytes, err := publicKey.ToByteArray()
		if err != nil {
			return nil, nil, err
		}

		blsKeysHex = append(blsKeysHex, hex.EncodeToString(publicKeyBytes))
	}

	return secretKeysBytes, blsKeysHex, nil
}
