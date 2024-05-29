package esdt

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
	issueWasmPath              = "../../../cmd/sovereignnode/chainSimulator/tests/testdata/issue.wasm"
	issuePrice                 = "5000000000000000000"
)

func TestSmartContract_IssueToken_MainChain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
		},
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

	nodeHandler := cs.GetNodeHandler(0)

	systemScAddress, err := chainSim.GetSysAccBytesAddress(nodeHandler)
	require.Nil(t, err)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(100)))
	require.Nil(t, err)
	nonce := uint64(0)

	deployedContractAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", issueWasmPath)
	deployedContractAddressBech32, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(deployedContractAddress)
	require.Nil(t, err)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, issueCost, "issue", uint64(60000000))

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.True(t, account.Balance == "0")

	issuedESDTs, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs(core.FungibleESDT)
	require.Nil(t, err)
	require.NotNil(t, issuedESDTs)
	require.True(t, len(issuedESDTs) == 1)
	tokenIdentifier := issuedESDTs[0]

	setRolesArgs := "setRoles@" + hex.EncodeToString([]byte(tokenIdentifier))
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, chainSim.ZeroValue, setRolesArgs, uint64(60000000))

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	esdtsRoles, _, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetESDTsRoles(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdtsRoles)
	require.True(t, len(esdtsRoles[tokenIdentifier]) == 3)
	require.Equal(t, core.ESDTRoleLocalMint, esdtsRoles[tokenIdentifier][0])
	require.Equal(t, core.ESDTRoleLocalBurn, esdtsRoles[tokenIdentifier][1])
	require.Equal(t, core.ESDTRoleTransfer, esdtsRoles[tokenIdentifier][2])

	expectedMintedAmount, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	mintTxArgs := "mint" +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(expectedMintedAmount.Bytes())
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, chainSim.ZeroValue, mintTxArgs, uint64(20000000))

	account, _, err = nodeHandler.GetFacadeHandler().GetAccount(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.True(t, account.Balance == "0")

	esdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotEmpty(t, esdts)
	require.Equal(t, expectedMintedAmount, esdts[issuedESDTs[0]].Value)
}
