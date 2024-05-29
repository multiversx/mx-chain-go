package esdt

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	issueWasmPath              = "../testdata/issue.wasm"
	issuePrice                 = "5000000000000000000"
)

// The test will deploy issue.wasm contract.
// The contract contains 3 endpoints (issue, setRoles and mint) which are called in the test
func TestSovereignChain_SmartContract_IssueToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
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
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	time.Sleep(time.Second) // wait for VM to be ready for processing queries

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

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

	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(deployedContractAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.True(t, account.Balance == "0")

	issuedESDTs, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs(core.FungibleESDT)
	require.Nil(t, err)
	require.NotNil(t, issuedESDTs)
	require.True(t, len(issuedESDTs) == 1)
	tokenIdentifier := issuedESDTs[0]

	setRolesArgs := "setRoles@" + hex.EncodeToString([]byte(tokenIdentifier))
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, chainSim.ZeroValue, setRolesArgs, uint64(60000000))

	checkAllRoles(t, nodeHandler, deployedContractAddressBech32, tokenIdentifier)

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

func checkAllRoles(t *testing.T, nodeHandler process.NodeHandler, address string, tokenIdentifier string) {
	esdtsRoles, _, err := nodeHandler.GetFacadeHandler().GetESDTsRoles(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdtsRoles)
	require.True(t, len(esdtsRoles[tokenIdentifier]) == 3)
	require.Equal(t, core.ESDTRoleLocalMint, esdtsRoles[tokenIdentifier][0])
	require.Equal(t, core.ESDTRoleLocalBurn, esdtsRoles[tokenIdentifier][1])
	require.Equal(t, core.ESDTRoleTransfer, esdtsRoles[tokenIdentifier][2])
}
