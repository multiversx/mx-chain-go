package vm

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/utils"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-core-go/core"
	api2 "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	adderWasmPath              = "../testdata/adder.wasm"
	issuePrice                 = "5000000000000000000"
)

var log = logger.GetOrCreate("issue-test")

func TestSmartContract_IssueToken(t *testing.T) {
	epochConfig, economicsConfig, sovereignExtraConfig, err := sovereignChainSimulator.LoadSovereignConfigs(sovereignConfigPath)
	require.Nil(t, err)

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignExtraConfig: *sovereignExtraConfig,
		ChainSimulatorArgs: chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
				cfg.EconomicsConfig = economicsConfig
				cfg.EpochConfig = epochConfig
				cfg.GeneralConfig.SovereignConfig = *sovereignExtraConfig
				cfg.GeneralConfig.VirtualMachine.Execution.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
				cfg.GeneralConfig.VirtualMachine.Querying.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(200)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	oneEgld := big.NewInt(1000000000000000000)
	initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(10))
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	// FIX THIS ????
	initialSupply := big.NewInt(144)
	issuedToken := issueESDT(t, cs, &nonce, wallet.Bytes, "TKN", initialSupply)
	require.True(t, strings.HasPrefix(issuedToken, "TKN-"))

	esdts, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs("FungibleESDT")
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, len(esdts) == 1)

	chLeaves := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}

	acc, _ := nodeHandler.GetStateComponents().AccountsAdapter().LoadAccount(vm.ESDTSCAddress)
	userAcc := acc.(state.UserAccountHandler)
	userAcc.GetAllLeaves(chLeaves, context.Background())

	for leaf := range chLeaves.LeavesChan {
		log.Error("KEY######", "leaf.Key", string(leaf.Key()))

	}
}

func issueESDT(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	nonce *uint64,
	sender []byte,
	tokenName string,
	supply *big.Int,
) string {
	// Step 2 - generate issue tx
	issueValue, _ := big.NewInt(0).SetString(issuePrice, 10)
	dataField := fmt.Sprintf("issue@%s@%s@%s@01@63616e55706772616465@74727565@63616e57697065@74727565@63616e467265657a65@74727565",
		hex.EncodeToString([]byte(tokenName)), hex.EncodeToString([]byte(tokenName)), hex.EncodeToString(supply.Bytes()))
	tx := &transaction.Transaction{
		Nonce:     *nonce,
		Value:     issueValue,
		SndAddr:   sender,
		RcvAddr:   vm.ESDTSCAddress,
		Data:      []byte(dataField),
		GasLimit:  100_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	issueTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 8)
	require.Nil(t, err)
	require.NotNil(t, issueTx)

	cs.GenerateBlocks(2)

	issuedTokens, err := cs.GetNodeHandler(0).GetFacadeHandler().GetAllIssuedESDTs("FungibleESDT")
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(issuedTokens), 1)

	for _, issuedToken := range issuedTokens {
		if strings.Contains(issuedToken, tokenName) {
			log.Info("issued token", "token", issuedToken)
			return issuedToken
		}
	}

	*nonce++
	require.Fail(t, "could not create token")
	return ""
}

func TestSmartContract_IssueToken_MainChain(t *testing.T) {
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 1,
		GenesisTimestamp:            time.Now().Unix(),
		RoundDurationInMillis:       uint64(6000),
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            2,
		MetaChainMinNodes:           2,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t",
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(5)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	oneEgld := big.NewInt(1000000000000000000)
	initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(10))
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	deployedContractAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "@10000000", adderWasmPath)

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum := big.NewInt(0).SetBytes(res.ReturnData[0]).Int64()
	require.Equal(t, 268435456, int(sum))

	issueCost := big.NewInt(5000000000000000000)
	tx1 := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, issueCost, "issue", uint64(60000000))
	require.False(t, string(tx1.Logs.Events[0].Topics[1]) == "sending value to non payable contract")

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	esdts, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs("FungibleESDT")
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, len(esdts) == 1)

	mintTxData := "mint@" + hex.EncodeToString([]byte(esdts[0]))
	tx2 := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, big.NewInt(0), mintTxData, uint64(20000000))
	require.NotNil(t, tx2)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	deployedAddrBech32, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(deployedContractAddress)
	require.Nil(t, err)
	_ = deployedAddrBech32

	expectedMintedAmount, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	esdtSC, _, err := cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens(deployedAddrBech32, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotEmpty(t, esdtSC)
	require.Equal(t, expectedMintedAmount, esdtSC[esdts[0]].Value)
}
