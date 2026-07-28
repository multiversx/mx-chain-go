package chainSimulator

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	simulatorAPI "github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorErrors "github.com/multiversx/mx-chain-go/node/chainSimulator/errors"
)

const consensusTestGenesisTimestamp = int64(1700000000)

var (
	consensusShortRoundsPerEpoch          = core.OptionalUint64{HasValue: true, Value: 10}
	consensusShortSupernovaRoundsPerEpoch = core.OptionalUint64{HasValue: true, Value: 20}
)

type consensusDriveNodeStub struct {
	subround   int
	generation uint64
}

func (stub *consensusDriveNodeStub) AdvanceConsensusClock() error {
	return nil
}

func (stub *consensusDriveNodeStub) RearmConsensusRound() error {
	return nil
}

func (stub *consensusDriveNodeStub) StepConsensusSubround() error {
	return nil
}

func (stub *consensusDriveNodeStub) WaitConsensusSubround() error {
	return nil
}

func (stub *consensusDriveNodeStub) ConsensusDriveState() (int, uint64, error) {
	return stub.subround, stub.generation, nil
}

// newConsensusTestSimulator builds an enable-consensus simulator over one shard plus metachain,
// with two single-key validator nodes per shard and a shared genesis timestamp.
func newConsensusTestSimulator(
	t *testing.T,
	roundsPerEpoch core.OptionalUint64,
	supernovaRoundsPerEpoch core.OptionalUint64,
) *simulator {
	return newConsensusTestSimulatorWithCrypto(
		t,
		roundsPerEpoch,
		supernovaRoundsPerEpoch,
		ConsensusModeBLS,
	)
}

func newConsensusTestSimulatorWithCrypto(
	t *testing.T,
	roundsPerEpoch core.OptionalUint64,
	supernovaRoundsPerEpoch core.OptionalUint64,
	consensusMode ConsensusMode,
) *simulator {
	sim, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    1,
		GenesisTimestamp:               consensusTestGenesisTimestamp,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 roundsPerEpoch,
		SupernovaRoundsPerEpoch:        supernovaRoundsPerEpoch,
		ApiInterface:                   simulatorAPI.NewNoApiInterface(),
		MinNodesPerShard:               2,
		MetaChainMinNodes:              2,
		ConsensusMode:                  consensusMode,
	})
	require.NoError(t, err)
	require.NotNil(t, sim)

	return sim
}

func TestChainSimulator_EnableConsensus_GenerateBlocks(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    1,
		GenesisTimestamp:               consensusTestGenesisTimestamp,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   simulatorAPI.NewNoApiInterface(),
		MinNodesPerShard:               1,
		MetaChainMinNodes:              1,
		ConsensusMode:                  ConsensusModeBLS,
	})
	require.NoError(t, err)
	require.NotNil(t, simulator)
	defer simulator.Close()

	require.NoError(t, simulator.GenerateBlocks(3))

	for _, shardID := range []uint32{0, core.MetachainShardId} {
		require.Len(t, simulator.consensusNodes[shardID], 1, "expected one node per validator in shard %d", shardID)

		header := simulator.GetNodeHandler(shardID).GetChainHandler().GetCurrentBlockHeader()
		require.False(t, check.IfNil(header), "shard %d did not commit any block", shardID)
		require.GreaterOrEqual(t, header.GetNonce(), uint64(1), "shard %d nonce did not advance", shardID)
	}
}

func TestChainSimulator_EnableFastConsensusCrypto_GeneratesQuorumBlocks(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    1,
		GenesisTimestamp:               consensusTestGenesisTimestamp,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   simulatorAPI.NewNoApiInterface(),
		MinNodesPerShard:               2,
		MetaChainMinNodes:              2,
		ConsensusMode:                  ConsensusModeFastCrypto,
	})
	require.NoError(t, err)
	require.NotNil(t, simulator)
	defer simulator.Close()

	require.NoError(t, simulator.GenerateBlocks(3))

	for _, shardID := range []uint32{0, core.MetachainShardId} {
		require.Len(t, simulator.consensusNodes[shardID], 2)

		header := simulator.GetNodeHandler(shardID).GetChainHandler().GetCurrentBlockHeader()
		require.False(t, check.IfNil(header))
		require.GreaterOrEqual(t, header.GetNonce(), uint64(1))
		require.Len(t, header.GetSignature(), 48)
		require.NotEmpty(t, header.GetPubKeysBitmap())
	}
}

func TestChainSimulator_InvalidConsensusMode(t *testing.T) {
	simulator, err := NewChainSimulator(ArgsChainSimulator{
		ConsensusMode: ConsensusModeFastCrypto + 1,
	})

	require.Nil(t, simulator)
	require.ErrorIs(t, err, chainSimulatorErrors.ErrInvalidConsensusMode)
}

func TestChainSimulator_EnableFastConsensusCrypto_ProducesAndVerifiesSupernovaProofs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := newConsensusTestSimulatorWithCrypto(
		t,
		consensusShortRoundsPerEpoch,
		consensusShortSupernovaRoundsPerEpoch,
		ConsensusModeFastCrypto,
	)
	defer simulator.Close()

	require.NoError(t, simulator.GenerateBlocksUntilEpochIsReached(3))
	require.NoError(t, simulator.GenerateBlocks(3))

	for _, shardID := range []uint32{0, core.MetachainShardId} {
		referenceHeader := simulator.GetNodeHandler(shardID).GetChainHandler().GetCurrentBlockHeader()
		require.True(t, referenceHeader.IsHeaderV3())
		require.GreaterOrEqual(t, referenceHeader.GetEpoch(), uint32(3))

		for nodeIndex, node := range simulator.consensusNodes[shardID] {
			nodeHeader := node.GetChainHandler().GetCurrentBlockHeader()
			require.Equal(t, referenceHeader.GetNonce(), nodeHeader.GetNonce())

			proof, err := node.GetDataComponents().Datapool().Proofs().GetProof(
				shardID,
				node.GetChainHandler().GetCurrentBlockHeaderHash(),
			)
			require.NoError(t, err, "shard %d node %d has no proof", shardID, nodeIndex)
			require.Len(t, proof.GetAggregatedSignature(), 48)
			require.NotEmpty(t, proof.GetPubKeysBitmap())
		}
	}
}

func TestLaggingConsensusDrivers_SelectsOnlyNodesBehindAfterRestart(t *testing.T) {
	advanced := &consensusDriveNodeStub{subround: 2, generation: 2}
	restarted := &consensusDriveNodeStub{subround: -1, generation: 3}
	partiallyCaughtUp := &consensusDriveNodeStub{subround: 1, generation: 3}
	drivers := []consensusDriveNode{advanced, restarted, partiallyCaughtUp}

	lagging, caughtUp, err := laggingConsensusDrivers(drivers)

	require.NoError(t, err)
	require.False(t, caughtUp)
	require.Equal(t, []consensusDriveNode{restarted, partiallyCaughtUp}, lagging)

	restarted.subround = 2
	partiallyCaughtUp.subround = 2
	lagging, caughtUp, err = laggingConsensusDrivers(drivers)
	require.NoError(t, err)
	require.True(t, caughtUp)
	require.Empty(t, lagging)
}

func TestUpdateConsensusDriveGenerations_DetectsRestart(t *testing.T) {
	first := &consensusDriveNodeStub{subround: 1, generation: 4}
	second := &consensusDriveNodeStub{subround: 1, generation: 5}
	drivers := []consensusDriveNode{first, second}
	generations := []uint64{4, 4}

	restarted, err := updateConsensusDriveGenerations(drivers, generations)

	require.NoError(t, err)
	require.True(t, restarted)
	require.Equal(t, []uint64{4, 5}, generations)
}

// TestEnableConsensus_ProducesBlocksAcrossEpochs verifies consensus remains live across Andromeda
// and Supernova activation and that every physical validator converges after the boundaries.
func TestEnableConsensus_ProducesBlocksAcrossEpochs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := newConsensusTestSimulator(t, consensusShortRoundsPerEpoch, consensusShortSupernovaRoundsPerEpoch)
	defer simulator.Close()

	const targetEpoch = int32(3)
	require.NoError(t, simulator.GenerateBlocksUntilEpochIsReached(targetEpoch))
	for shardID, nodeHandler := range simulator.nodes {
		header := nodeHandler.GetChainHandler().GetCurrentBlockHeader()
		require.False(t, check.IfNil(header), "shard %d did not commit any block", shardID)
		require.GreaterOrEqual(t, int32(header.GetEpoch()), targetEpoch,
			"GenerateBlocksUntilEpochIsReached returned before shard %d committed epoch %d", shardID, targetEpoch)
	}
	require.NoError(t, simulator.GenerateBlocks(3))

	for _, shardID := range []uint32{0, core.MetachainShardId} {
		nodeHandler := simulator.GetNodeHandler(shardID)
		enableEpochsHandler := nodeHandler.GetCoreComponents().EnableEpochsHandler()
		require.GreaterOrEqual(t, int32(enableEpochsHandler.GetCurrentEpoch()), targetEpoch,
			"shard %d did not reach epoch %d", shardID, targetEpoch)
		require.True(t, enableEpochsHandler.IsFlagEnabled(common.AndromedaFlag),
			"shard %d must run under Andromeda at epoch %d", shardID, targetEpoch)
		require.True(t, enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag),
			"shard %d must run under Supernova at epoch %d", shardID, targetEpoch)

		header := nodeHandler.GetChainHandler().GetCurrentBlockHeader()
		require.False(t, check.IfNil(header), "shard %d did not commit any block", shardID)
		require.GreaterOrEqual(t, int32(header.GetEpoch()), targetEpoch,
			"shard %d current block is still in epoch %d", shardID, header.GetEpoch())
		require.GreaterOrEqual(t, header.GetNonce(), uint64(30),
			"shard %d committed too few blocks to demonstrate liveness", shardID)

		for idx, node := range simulator.consensusNodes[shardID] {
			nodeHeader := node.GetChainHandler().GetCurrentBlockHeader()
			require.False(t, check.IfNil(nodeHeader), "shard %d node %d did not commit any block", shardID, idx)
			require.Equal(t, header.GetNonce(), nodeHeader.GetNonce(),
				"shard %d node %d did not converge after the epoch boundaries", shardID, idx)
			require.GreaterOrEqual(t, int32(nodeHeader.GetEpoch()), targetEpoch,
				"shard %d node %d is stuck before epoch %d", shardID, idx, targetEpoch)
		}
	}
}

func TestEnableConsensus_InjectedStateSurvivesDeferredExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := newConsensusTestSimulator(t, consensusShortRoundsPerEpoch, consensusShortSupernovaRoundsPerEpoch)
	defer simulator.Close()

	require.NoError(t, simulator.GenerateBlocksUntilEpochIsReached(3))

	const (
		address = "erd1yf9z866ee645k93ypk9t98njyakytmksynjpa43tjdq9dhx87cdq6w658w"
		balance = "1000000000000000000"
	)
	require.NoError(t, simulator.SetStateMultiple([]*dtos.AddressState{
		{
			Address: address,
			Balance: balance,
		},
	}))

	for round := 0; round < 5; round++ {
		require.NoError(t, simulator.GenerateBlocks(1))

		for nodeIndex, node := range simulator.consensusNodes[0] {
			account, _, err := node.GetFacadeHandler().GetAccount(address, coreAPI.AccountQueryOptions{})
			require.NoError(t, err)
			require.Equal(t, balance, account.Balance,
				"node %d lost simulator-injected state after Supernova round %d", nodeIndex, round+1)
		}
	}
}

func TestEnableConsensus_NodesConverge(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := newConsensusTestSimulator(t, consensusShortRoundsPerEpoch, consensusShortSupernovaRoundsPerEpoch)
	defer simulator.Close()

	require.NoError(t, simulator.GenerateBlocks(15))

	for _, shardID := range []uint32{0, core.MetachainShardId} {
		consensusNodes := simulator.consensusNodes[shardID]
		require.Len(t, consensusNodes, 2, "expected one node per validator in shard %d", shardID)

		referenceChain := consensusNodes[0].GetChainHandler()
		referenceHeader := referenceChain.GetCurrentBlockHeader()
		require.False(t, check.IfNil(referenceHeader), "shard %d node 0 did not commit any block", shardID)
		expectedMinNonce := uint64(15)
		if shardID == core.MetachainShardId {
			// The first post-epoch-start metachain round can legitimately be empty while
			// the consensus version restarts. No other requested round should be lost.
			expectedMinNonce--
		}
		require.GreaterOrEqual(t, referenceHeader.GetNonce(), expectedMinNonce,
			"shard %d lost more than the one expected transition round", shardID)

		for idx, node := range consensusNodes[1:] {
			chain := node.GetChainHandler()
			header := chain.GetCurrentBlockHeader()
			require.False(t, check.IfNil(header), "shard %d node %d did not commit any block", shardID, idx+1)
			require.Equal(t, referenceHeader.GetNonce(), header.GetNonce(),
				"shard %d node %d sits at a different height than node 0", shardID, idx+1)
			require.Equal(t,
				hex.EncodeToString(referenceChain.GetCurrentBlockHeaderHash()),
				hex.EncodeToString(chain.GetCurrentBlockHeaderHash()),
				"shard %d node %d committed a different block than node 0 at nonce %d",
				shardID, idx+1, header.GetNonce())
		}
	}
}

func TestEnableConsensus_WaitingValidatorsRemainLiveAcrossShuffles(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    1,
		GenesisTimestamp:               consensusTestGenesisTimestamp,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 consensusShortRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        consensusShortSupernovaRoundsPerEpoch,
		ApiInterface:                   simulatorAPI.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListShard:       2,
		NumNodesWaitingListMeta:        2,
		ConsensusMode:                  ConsensusModeBLS,
	})
	require.NoError(t, err)
	defer simulator.Close()

	for _, shardID := range []uint32{0, core.MetachainShardId} {
		require.Len(t, simulator.consensusNodes[shardID], 5)
	}

	require.NoError(t, simulator.GenerateBlocksUntilEpochIsReached(4))
	require.NoError(t, simulator.GenerateBlocks(5))

	for _, shardID := range []uint32{0, core.MetachainShardId} {
		reference := simulator.consensusNodes[shardID][0].GetChainHandler().GetCurrentBlockHeader()
		require.False(t, check.IfNil(reference))
		require.GreaterOrEqual(t, reference.GetEpoch(), uint32(4))

		for idx, node := range simulator.consensusNodes[shardID][1:] {
			header := node.GetChainHandler().GetCurrentBlockHeader()
			require.False(t, check.IfNil(header), "shard %d node %d has no committed block", shardID, idx+1)
			require.Equal(t, reference.GetNonce(), header.GetNonce(),
				"shard %d node %d did not remain caught up across validator shuffles", shardID, idx+1)
		}
	}
}
