package epochChange

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	proc "github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	testsFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
)

var log = logger.GetOrCreate("epoch-change")

func TestSovereignChainSimulator_EpochChange(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    50, // do not lower this value so that each validator can participate in consensus as leader to get rewards
	}

	sovConfig := config.SovereignConfig{}

	sovRequestHandler := &testscommon.ExtendedShardHeaderRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
				require.Fail(t, "should not request miniBlock")
			},
			RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblocksHashes [][]byte) {
				require.Fail(t, "should not request miniBlocks")
			},
			RequestRewardTxHandlerCalled: func(destShardID uint32, txHashes [][]byte) {
				require.Fail(t, "should not request reward txs")
			},
		},
	}
	sovRequestHandlerFactory := &testsFactory.RequestHandlerFactoryMock{
		CreateRequestHandlerCalled: func(args requestHandlers.RequestHandlerArgs) (proc.RequestHandler, error) {
			return sovRequestHandler, nil
		},
	}

	var protocolSustainabilityAddress string
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   true,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    uint64(6000),
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         6,
			NumNodesWaitingListShard: 2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				newCfg := config.EnableEpochs{}
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
				newCfg.BLSMultiSignerEnableEpoch = cfg.EpochConfig.EnableEpochs.BLSMultiSignerEnableEpoch
				newCfg.MaxNodesChangeEnableEpoch = []config.MaxNodesChangeConfig{
					{
						EpochEnable:            0,
						MaxNumNodes:            8,
						NodesToShufflePerShard: 2,
					},
				}

				protocolSustainabilityAddress = cfg.EconomicsConfig.RewardsSettings.RewardsConfigByEpoch[0].ProtocolSustainabilityAddress
				cfg.EpochConfig.EnableEpochs = newCfg
				sovConfig = cfg.GeneralConfig.SovereignConfig
			},
			CreateRunTypeComponents: func(args runType.ArgsRunTypeComponents) (factory.RunTypeComponentsHolder, error) {
				runTypeComps, err := common.CreateSovereignRunTypeComponents(args, sovConfig)
				require.Nil(t, err)

				runTypeCompsHolder := components.GetRunTypeComponentsStub(runTypeComps)
				runTypeCompsHolder.RequestHandlerFactory = sovRequestHandlerFactory

				return runTypeCompsHolder, nil
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	nonce := uint64(0)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	protocolSustainabilityAddrBalance, _, err := nodeHandler.GetFacadeHandler().GetBalance(protocolSustainabilityAddress, apiData.AccountQueryOptions{})
	require.Nil(t, err)
	require.Empty(t, protocolSustainabilityAddrBalance.Bytes())

	trie := nodeHandler.GetStateComponents().TriesContainer().Get([]byte(dataRetriever.PeerAccountsUnit.String()))
	require.NotNil(t, trie)

	// Generate enough blocks so that we achieve > 1500 trie storage reads (from MaxNumberOfTrieReadsPerTx gasSchedule cfg)
	err = cs.GenerateBlocksUntilEpochIsReached(45)
	require.Nil(t, err)
	require.Equal(t, uint32(45), nodeHandler.GetCoreComponents().EpochNotifier().CurrentEpoch())

	accFeesInEpoch, devFeesInEpoch := getAllFeesInEpoch(nodeHandler)
	require.Empty(t, accFeesInEpoch.Bytes())
	require.Empty(t, devFeesInEpoch.Bytes())

	staking.StakeNodes(t, cs, nodeHandler, 10)
	err = nodeHandler.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)

	auctionList, err := nodeHandler.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)
	require.Len(t, auctionList, 10)

	validators := nodeHandler.GetProcessComponents().ValidatorsProvider().GetLatestValidators()
	require.Len(t, validators, 18)

	// check here fees increase after stake txs
	accFeesInEpoch, devFeesInEpoch = getAllFeesInEpoch(nodeHandler)
	require.NotEmpty(t, accFeesInEpoch)
	require.NotEmpty(t, devFeesInEpoch)

	currentEpoch := nodeHandler.GetCoreComponents().EpochNotifier().CurrentEpoch()
	for epoch := currentEpoch + 1; epoch < currentEpoch+6; epoch++ {
		allOwnersBalance := getConsensusOwnersBalances(t, nodeHandler)

		err = cs.GenerateBlocksUntilEpochIsReached(int32(epoch))
		require.Nil(t, err)

		checkEpochChangeHeader(t, nodeHandler, allOwnersBalance, protocolSustainabilityAddress)
		requireValidatorBalancesIncreasedAfterRewards(t, nodeHandler, allOwnersBalance)
		checkProtocolSustainabilityAddressBalanceIncreased(t, nodeHandler, protocolSustainabilityAddress, protocolSustainabilityAddrBalance)

		qualified, unqualified := staking.GetQualifiedAndUnqualifiedNodes(t, nodeHandler)
		require.Len(t, qualified, 2)
		require.Len(t, unqualified, 8)

		chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, wallet.Bytes, chainSim.ZeroValue, "data", uint64(10000000))

		accFeesInEpoch, devFeesInEpoch = getAllFeesInEpoch(nodeHandler)
		require.NotEmpty(t, accFeesInEpoch)
		require.Empty(t, devFeesInEpoch)

		accFeesTotal, devFeesTotal := getAllFees(nodeHandler)
		require.NotEmpty(t, accFeesTotal.Bytes())
		require.Empty(t, devFeesTotal.Bytes())
	}
}

func checkEpochChangeHeader(
	t *testing.T,
	nodeHandler process.NodeHandler,
	allOwnersBalance map[string]*big.Int,
	protocolSustainabilityAddress string,
) {
	currentHeader := nodeHandler.GetDataComponents().Blockchain().GetCurrentBlockHeader()
	require.True(t, currentHeader.IsStartOfEpochBlock())

	mbs := currentHeader.GetMiniBlockHeaderHandlers()
	require.Len(t, mbs, 2)

	require.Equal(t, block.RewardsBlock, block.Type(mbs[0].GetTypeInt32()))
	require.Equal(t, block.PeerBlock, block.Type(mbs[1].GetTypeInt32()))

	require.Equal(t, mbs[0].GetTxCount(), uint32(7))  // consensus group reward txs = 6 + 1 reward tx protocol sustainability
	require.Equal(t, mbs[1].GetTxCount(), uint32(18)) // 18 validators in total => 18 peer block updates

	require.Equal(t, uint32(25), currentHeader.GetTxCount())

	unComputedRootHash := nodeHandler.GetCoreComponents().Hasher().Compute("uncomputed root hash")
	require.NotEqual(t, unComputedRootHash, currentHeader.GetRootHash())
	require.NotEqual(t, unComputedRootHash, currentHeader.GetValidatorStatsRootHash())
	require.NotNil(t, currentHeader.GetReceiptsHash())

	rootHash, err := nodeHandler.GetStateComponents().AccountsAdapter().RootHash()
	require.Nil(t, err)
	require.Equal(t, rootHash, currentHeader.GetRootHash())

	validatorRootHash, err := nodeHandler.GetStateComponents().PeerAccounts().RootHash()
	require.Nil(t, err)
	require.Equal(t, validatorRootHash, currentHeader.GetValidatorStatsRootHash())

	checkEpochChangeRewardsMB(t, nodeHandler, mbs[0], currentHeader, allOwnersBalance, protocolSustainabilityAddress)
}

func checkEpochChangeRewardsMB(
	t *testing.T,
	nodeHandler process.NodeHandler,
	mb data.MiniBlockHeaderHandler,
	currentHeader data.HeaderHandler,
	allOwnersBalance map[string]*big.Int,
	protocolSustainabilityAddress string,
) {
	mbRewardBytes, ok := nodeHandler.GetDataComponents().Datapool().MiniBlocks().Get(mb.GetHash())
	require.True(t, ok)

	mbReward, castOk := mbRewardBytes.(*block.MiniBlock)
	require.True(t, castOk)
	require.Len(t, mbReward.TxHashes, 7)

	owners := getOwnersMap(allOwnersBalance, nodeHandler.GetCoreComponents().AddressPubKeyConverter())
	owners[protocolSustainabilityAddress] = struct{}{}
	for _, txHash := range mbReward.TxHashes {
		tx, err := nodeHandler.GetFacadeHandler().GetTransaction(hex.EncodeToString(txHash), false)
		require.Nil(t, err)

		require.Equal(t, string(transaction.TxTypeReward), tx.Type)
		require.Equal(t, currentHeader.GetRound(), tx.Round)
		require.Equal(t, currentHeader.GetEpoch(), tx.Epoch)
		require.Equal(t, "sovereign", tx.Sender)
		require.Equal(t, core.SovereignChainShardId, tx.SourceShard)

		_, found := owners[tx.Receiver]
		require.True(t, found)
		delete(owners, tx.Receiver)
	}

	require.Empty(t, owners)
}

func getOwnersMap(allOwnersBalance map[string]*big.Int, pkConv core.PubkeyConverter) map[string]struct{} {
	ret := make(map[string]struct{})
	for owner := range allOwnersBalance {
		encodedKey, _ := pkConv.Encode([]byte(owner))
		ret[encodedKey] = struct{}{}
	}

	return ret
}

func getConsensusOwnersBalances(t *testing.T, nodeHandler process.NodeHandler) map[string]*big.Int {
	currentHeader := nodeHandler.GetDataComponents().Blockchain().GetCurrentBlockHeader()
	nodesCoordinator := nodeHandler.GetProcessComponents().NodesCoordinator()

	validators, err := headerCheck.ComputeConsensusGroup(currentHeader, nodesCoordinator)
	require.Nil(t, err)
	require.Len(t, validators, nodesCoordinator.ConsensusGroupSize(core.SovereignChainShardId))

	allOwnersBalance := make(map[string]*big.Int)
	for _, validator := range validators {
		owner := staking.GetBLSKeyOwner(t, nodeHandler, validator.PubKey())

		acc, err := nodeHandler.GetStateComponents().AccountsAdapter().GetExistingAccount(owner)
		require.Nil(t, err)

		userAcc, castOk := acc.(data.UserAccountHandler)
		require.True(t, castOk)

		allOwnersBalance[string(owner)] = userAcc.GetBalance()
	}

	return allOwnersBalance
}

func requireValidatorBalancesIncreasedAfterRewards(
	t *testing.T,
	nodeHandler process.NodeHandler,
	ownersBalance map[string]*big.Int,
) {
	require.NotEmpty(t, ownersBalance)
	for owner, previousBalance := range ownersBalance {
		acc, err := nodeHandler.GetStateComponents().AccountsAdapter().GetExistingAccount([]byte(owner))
		require.Nil(t, err)

		userAcc, castOk := acc.(data.UserAccountHandler)
		require.True(t, castOk)

		currentBalance := userAcc.GetBalance()

		log.Info("checking validator owners balance after rewards",
			"owner", hex.EncodeToString([]byte(owner)),
			"previous balance", previousBalance.String(),
			"current balance", currentBalance.String())

		require.True(t, currentBalance.Cmp(previousBalance) > 0)
	}
}

func checkProtocolSustainabilityAddressBalanceIncreased(
	t *testing.T,
	nodeHandler process.NodeHandler,
	protocolSustainabilityAddress string,
	previousBalance *big.Int,
) {
	currBalance, _, err := nodeHandler.GetFacadeHandler().GetBalance(protocolSustainabilityAddress, apiData.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotEmpty(t, currBalance.String())

	require.True(t, currBalance.Cmp(previousBalance) > 0)
}

func getAllFeesInEpoch(nodeHandler process.NodeHandler) (*big.Int, *big.Int) {
	sovHdr := common.GetCurrentSovereignHeader(nodeHandler)
	return sovHdr.GetAccumulatedFeesInEpoch(), sovHdr.GetDevFeesInEpoch()
}

func getAllFees(nodeHandler process.NodeHandler) (*big.Int, *big.Int) {
	sovHdr := common.GetCurrentSovereignHeader(nodeHandler)
	return sovHdr.GetAccumulatedFees(), sovHdr.GetDeveloperFees()
}
