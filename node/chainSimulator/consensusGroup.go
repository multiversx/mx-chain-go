package chainSimulator

import (
	"bytes"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

// consensusKeyIndexes returns the indexes, within the ordered ValidatorsPrivateKeys slice, of the
// initially eligible validators assigned to the shard at the given creation index (-1 = metachain).
// These keys get dedicated physical nodes. Waiting validators are created separately after all
// initially eligible nodes.
func consensusKeyIndexes(args ArgsBaseChainSimulator, idx int) []int {
	if !args.ConsensusMode.isEnabled() {
		return nil
	}

	indexes := make([]int, 0)
	if idx == -1 {
		for keyIdx := 0; keyIdx < int(args.MetaChainMinNodes); keyIdx++ {
			indexes = append(indexes, keyIdx)
		}
	} else {
		// NodesSetup assigns the eligible validators FIFO: metachain first, then each
		// shard. Waiting validators are assigned only after all eligible validators.
		baseIndex := int(args.MetaChainMinNodes) + idx*int(args.MinNodesPerShard)
		for keyIdx := baseIndex; keyIdx < baseIndex+int(args.MinNodesPerShard); keyIdx++ {
			indexes = append(indexes, keyIdx)
		}
	}

	return indexes
}

// createEligibleNodes keeps direct-mode construction unchanged and, in consensus mode, constructs
// the metachain and shard node groups concurrently. Nodes inside each shard are also independent:
// every constructor receives cloned mutable configs and owns its storage/state. Build those nodes
// concurrently, then publish the indexed result slice in validator-key order.
//
// A node owns its storage, state, component holders and genesis processing. The only objects shared
// between construction lanes are the synchronized in-memory broadcast network and heartbeat
// monitor. Building one lane per shard therefore removes the dominant serial startup cost without
// changing validator topology, key assignment or consensus behavior.
func (s *simulator) createEligibleNodes(
	outputConfigs configs.ArgsConfigsSimulator,
	args ArgsBaseChainSimulator,
	monitor simulatorHeartbeatMonitor,
) error {
	if !args.ConsensusMode.isEnabled() {
		for idx := -1; idx < int(args.NumOfShards); idx++ {
			shardIDStr := fmt.Sprintf("%d", idx)
			if idx == -1 {
				shardIDStr = "metachain"
			}

			err := s.createShardNodes(idx, shardIDStr, outputConfigs, args, monitor)
			if err != nil {
				return err
			}
		}

		return nil
	}

	numGroups := int(args.NumOfShards) + 1
	errs := make([]error, numGroups)
	wg := sync.WaitGroup{}
	wg.Add(numGroups)

	for idx := -1; idx < int(args.NumOfShards); idx++ {
		idx := idx
		resultIndex := idx + 1
		shardIDStr := fmt.Sprintf("%d", idx)
		if idx == -1 {
			shardIDStr = "metachain"
		}

		go func() {
			defer wg.Done()
			errs[resultIndex] = s.createShardNodes(idx, shardIDStr, outputConfigs, args, monitor)
		}()
	}

	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	// Concurrent construction can finish in any order. Restore the historical deterministic
	// metachain, shard-0, shard-1, ... handler order used by post-round catch-up.
	sort.Slice(s.handlers, func(i, j int) bool {
		return shardHandlerOrder(s.handlers[i].shardID) < shardHandlerOrder(s.handlers[j].shardID)
	})

	return nil
}

func shardHandlerOrder(shardID uint32) uint64 {
	if shardID == core.MetachainShardId {
		return 0
	}

	return uint64(shardID) + 1
}

type waitingNodeSpec struct {
	waitingIdx int
}

// createShardNodes creates every node of the shard at the given creation index and runs the
// per-node genesis setup. With a single node it behaves exactly as before; with a consensus group
// of N it spawns N single-key nodes (distinct peer IDs and one managed validator key each), records
// the first as the shard's primary (used for queries and the direct-mode block creator) and all of
// them in consensusNodes (driven together each round).
func (s *simulator) createShardNodes(
	idx int,
	shardIDStr string,
	outputConfigs configs.ArgsConfigsSimulator,
	args ArgsBaseChainSimulator,
	monitor simulatorHeartbeatMonitor,
) error {
	keyIndexes := consensusKeyIndexes(args, idx)
	numNodes := 1
	if args.ConsensusMode.isEnabled() {
		numNodes = len(keyIndexes)
	}

	results := make([]process.NodeHandler, numNodes)
	errs := make([]error, numNodes)
	wg := sync.WaitGroup{}
	wg.Add(numNodes)
	for nodeIdx := 0; nodeIdx < numNodes; nodeIdx++ {
		nodeIdx := nodeIdx
		go func() {
			defer wg.Done()

			pemOverride := ""
			if args.ConsensusMode.isEnabled() {
				keyIdx := keyIndexes[nodeIdx]
				if keyIdx >= len(outputConfigs.ValidatorsPrivateKeys) {
					errs[nodeIdx] = fmt.Errorf("%w: validator key index %d out of range (have %d keys)",
						errShardSetupError, keyIdx, len(outputConfigs.ValidatorsPrivateKeys))
					return
				}

				pemName := fmt.Sprintf("shard-%s-node-%d.pem", shardIDStr, nodeIdx)
				var err error
				pemOverride, err = writeSingleKeyPem(
					filepath.Join(args.TempDir, "consensus-keys"),
					pemName,
					outputConfigs.ValidatorsPrivateKeys[keyIdx],
				)
				if err != nil {
					errs[nodeIdx] = err
					return
				}
			}

			node, err := s.createTestNodeWithKeys(outputConfigs, args, shardIDStr, monitor, pemOverride)
			if err != nil {
				errs[nodeIdx] = err
				return
			}

			err = setupNodeGenesis(node, args)
			if err != nil {
				errs[nodeIdx] = err
				return
			}

			results[nodeIdx] = node
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	shardID := results[0].GetShardCoordinator().SelfId()
	chainHandler, err := process.NewBlocksCreator(
		results[0],
		monitor,
		args.CreateBlockMaxTimePercent,
		args.BypassCreateBlockTimeCheck,
	)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	for nodeIdx, node := range results {
		if node.GetShardCoordinator().SelfId() != shardID {
			return fmt.Errorf("%w: eligible validator %d expected shard %d, got %d",
				errShardSetupError, nodeIdx, shardID, node.GetShardCoordinator().SelfId())
		}
		if nodeIdx == 0 {
			s.nodes[shardID] = node
			s.handlers = append(s.handlers, shardChainHandler{shardID: shardID, handler: chainHandler})
		}
		s.consensusNodes[shardID] = append(s.consensusNodes[shardID], node)
	}

	return nil
}

// waitingShardAssignment mirrors NodesSetup's round-robin assignment of the keys after the initial
// eligible block. The metachain ID wraps the next assignment back to shard 0.
func waitingShardAssignment(numWaiting int, numShards uint32) []uint32 {
	shards := make([]uint32, numWaiting)
	currentShard := uint32(0)
	for idx := 0; idx < numWaiting; idx++ {
		currentShard = (currentShard + 1) % (numShards + 1)
		if currentShard == numShards {
			currentShard = core.MetachainShardId
		}
		shards[idx] = currentShard
	}

	return shards
}

// createWaitingNodes creates a physical single-key node for every waiting validator in consensus
// mode. Consensus-mode shuffling is kept within a shard (see CreateCoreComponents), so driving and
// catching up these nodes each round lets them replace demoted validators without losing quorum.
func (s *simulator) createWaitingNodes(
	outputConfigs configs.ArgsConfigsSimulator,
	args ArgsBaseChainSimulator,
	monitor simulatorHeartbeatMonitor,
) error {
	if !args.ConsensusMode.isEnabled() {
		return nil
	}

	eligibleTotal := int(args.MetaChainMinNodes) + int(args.NumOfShards)*int(args.MinNodesPerShard)
	numWaiting := len(outputConfigs.ValidatorsPrivateKeys) - eligibleTotal
	if numWaiting <= 0 {
		return nil
	}

	specsByShard := make(map[uint32][]waitingNodeSpec, args.NumOfShards+1)
	shards := waitingShardAssignment(numWaiting, args.NumOfShards)
	for waitingIdx, shardID := range shards {
		specsByShard[shardID] = append(specsByShard[shardID], waitingNodeSpec{
			waitingIdx: waitingIdx,
		})
	}

	shardIDs := make([]uint32, 0, args.NumOfShards+1)
	shardIDs = append(shardIDs, core.MetachainShardId)
	for shardID := uint32(0); shardID < args.NumOfShards; shardID++ {
		shardIDs = append(shardIDs, shardID)
	}

	errs := make([]error, len(shardIDs))
	wg := sync.WaitGroup{}
	wg.Add(len(shardIDs))
	for resultIndex, shardID := range shardIDs {
		resultIndex := resultIndex
		shardID := shardID

		go func() {
			defer wg.Done()
			errs[resultIndex] = s.createWaitingNodesForShard(
				outputConfigs,
				args,
				monitor,
				eligibleTotal,
				shardID,
				specsByShard[shardID],
			)
		}()
	}

	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *simulator) createWaitingNodesForShard(
	outputConfigs configs.ArgsConfigsSimulator,
	args ArgsBaseChainSimulator,
	monitor simulatorHeartbeatMonitor,
	eligibleTotal int,
	shardID uint32,
	specs []waitingNodeSpec,
) error {
	s.mutex.RLock()
	firstNodeIndex := len(s.consensusNodes[shardID])
	s.mutex.RUnlock()

	shardIDStr := "metachain"
	if shardID != core.MetachainShardId {
		shardIDStr = fmt.Sprintf("%d", shardID)
	}

	pemOverrides := make([]string, len(specs))
	for localIndex, spec := range specs {
		waitingIdx := spec.waitingIdx
		nodeIdx := firstNodeIndex + localIndex
		pemName := fmt.Sprintf("shard-%s-node-%d.pem", shardIDStr, nodeIdx)
		pemOverride, err := writeSingleKeyPem(
			filepath.Join(args.TempDir, "consensus-keys"),
			pemName,
			outputConfigs.ValidatorsPrivateKeys[eligibleTotal+waitingIdx],
		)
		if err != nil {
			return err
		}
		pemOverrides[localIndex] = pemOverride
	}

	results := make([]process.NodeHandler, len(specs))
	errs := make([]error, len(specs))
	wg := sync.WaitGroup{}
	wg.Add(len(specs))
	for localIndex := range specs {
		localIndex := localIndex
		go func() {
			defer wg.Done()

			node, err := s.createTestNodeWithKeys(
				outputConfigs,
				args,
				shardIDStr,
				monitor,
				pemOverrides[localIndex],
			)
			if err != nil {
				errs[localIndex] = err
				return
			}
			if node.GetShardCoordinator().SelfId() != shardID {
				errs[localIndex] = fmt.Errorf("%w: waiting validator %d expected shard %s, got %d",
					errShardSetupError, specs[localIndex].waitingIdx, shardIDStr, node.GetShardCoordinator().SelfId())
				return
			}

			err = setupNodeGenesis(node, args)
			if err != nil {
				errs[localIndex] = err
				return
			}
			results[localIndex] = node
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	s.mutex.Lock()
	for _, node := range results {
		s.consensusNodes[shardID] = append(s.consensusNodes[shardID], node)
	}
	s.mutex.Unlock()

	return nil
}

// setupNodeGenesis runs the per-node genesis initialization: the metachain processes its system-SC
// genesis, and every node is given its epoch-start header for the blockchain hook, its stored genesis
// header and an initialized system account.
func setupNodeGenesis(node process.NodeHandler, args ArgsBaseChainSimulator) error {
	var epochStartBlockHeader data.HeaderHandler

	shardID := node.GetShardCoordinator().SelfId()
	if shardID == core.MetachainShardId {
		err := initializeMetachainGenesisState(node)
		if err != nil {
			return err
		}

		epochStartBlockHeader = &block.MetaBlock{
			Nonce:     args.InitialNonce,
			Epoch:     args.InitialEpoch,
			Round:     uint64(args.InitialRound),
			TimeStamp: uint64(node.GetCoreComponents().RoundHandler().TimeStamp().Unix()),
		}
	} else {
		epochStartBlockHeader = &block.MetaBlockV3{
			Nonce:       args.InitialNonce,
			Epoch:       args.InitialEpoch,
			Round:       uint64(args.InitialRound),
			TimestampMs: uint64(node.GetCoreComponents().RoundHandler().TimeStamp().UnixMilli()),
		}
	}

	genesisBlock := node.GetDataComponents().Blockchain().GetGenesisHeader()
	err := node.GetDataComponents().Datapool().Transactions().OnExecutedBlock(genesisBlock, genesisBlock.GetRootHash())
	if err != nil {
		return err
	}

	err = node.GetProcessComponents().BlockchainHook().SetEpochStartHeader(epochStartBlockHeader)
	if err != nil {
		return err
	}

	headerBytes, err := node.GetCoreComponents().InternalMarshalizer().Marshal(genesisBlock)
	if err != nil {
		return err
	}

	storer, err := node.GetDataComponents().StorageService().GetStorer(dataRetriever.GetHeadersDataUnit(shardID))
	if err != nil {
		return err
	}
	identifier := []byte(core.EpochStartIdentifier(args.InitialEpoch))
	err = storer.Put(identifier, headerBytes)
	if err != nil {
		return err
	}

	return node.SetKeyValueForAddress(core.SystemAccountAddress, make(map[string]string))
}

// writeSingleKeyPem writes a one-key validator PEM in the format the crypto components' key loader
// expects and returns the file path
func writeSingleKeyPem(dir string, name string, sk crypto.PrivateKey) (string, error) {
	converter, err := pubkeyConverter.NewHexPubkeyConverter(96)
	if err != nil {
		return "", err
	}

	pkBytes, err := sk.GeneratePublic().ToByteArray()
	if err != nil {
		return "", err
	}
	pkString, err := converter.Encode(pkBytes)
	if err != nil {
		return "", err
	}
	skBytes, err := sk.ToByteArray()
	if err != nil {
		return "", err
	}

	buff := bytes.Buffer{}
	blk := pem.Block{
		Type:  "PRIVATE KEY for " + pkString,
		Bytes: []byte(hex.EncodeToString(skBytes)),
	}
	if err = pem.Encode(&buff, &blk); err != nil {
		return "", err
	}

	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	pemPath := filepath.Join(dir, name)
	if err = os.WriteFile(pemPath, buff.Bytes(), 0644); err != nil {
		return "", err
	}

	return pemPath, nil
}
