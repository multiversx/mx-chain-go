package delegation

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
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
)

func TestSovereignChainSimulator_NewDelegationSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.DelegationSmartContractEnableEpoch = 0
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(2500)))
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txData := "createNewDelegationContract" +
		"@" + hex.EncodeToString(big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(100000)).Bytes()) +
		"@64"
	cost := big.NewInt(0).Mul(chainSim.OneEGLD, big.NewInt(1250))
	txResult := chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.DelegationManagerSCAddress, cost, txData, uint64(60000000))
	chainSim.RequireSuccessfulTransaction(t, txResult)

	secondDelegationSCAddress := txResult.Logs.Events[1].Topics[4]
	secondDelegationSCAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(secondDelegationSCAddress)
	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(secondDelegationSCAddressBech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)
	require.True(t, len(account.Code) > 0)
}
