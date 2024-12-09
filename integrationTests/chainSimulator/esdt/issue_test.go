package esdt

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
	issuePrice                 = "5000000000000000000"
	tokenPrefix                = "sov1"
)

func TestChainSimulator_IssueESDTWithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              3,
		GenesisTimestamp:         time.Now().Unix(),
		RoundDurationInMillis:    uint64(6000),
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.ESDTSystemSCConfig.ESDTPrefix = tokenPrefix
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	// Step 1 - set an initial balance for the address that will initialize all the transactions
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
	})
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// Step 2 - generate issue tx
	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	initialSupply, _ := big.NewInt(0).SetString("144", 10)
	tokenName := "Token"
	tokenTicker := "TKN"
	numDecimals := 18
	issuedToken := chainSim.IssueFungible(t, cs, initialAddrBytes, &nonce, issueCost, tokenName, tokenTicker, numDecimals, initialSupply)
	require.True(t, strings.HasPrefix(issuedToken, tokenPrefix+"-"+tokenTicker+"-"))

	// Step 3 - send issued esdt
	receiver := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	receiverBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(receiver)
	receivedTokens := big.NewInt(11)
	require.Nil(t, err)
	chainSim.TransferESDT(t, cs, initialAddrBytes, receiverBytes, &nonce, issuedToken, receivedTokens)

	// Step 4 - check balances
	chainSim.RequireAccountHasToken(t, cs, issuedToken, receiver, receivedTokens)
	chainSim.RequireAccountHasToken(t, cs, issuedToken, initialAddress, big.NewInt(0).Sub(initialSupply, receivedTokens))
}
