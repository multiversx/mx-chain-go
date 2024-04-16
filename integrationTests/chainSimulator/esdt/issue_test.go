package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	dataApi "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
	issuePrice                 = "5000000000000000000"
)

var log = logger.GetOrCreate("issue-tests")

func TestChainSimulator_IssueESDTWithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(1)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   false,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              numOfShards,
		GenesisTimestamp:         startTime,
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8 // 8 nodes until new nodes will be placed on queue
			configs.SetMaxNumberOfNodesInConfigs(cfg, newNumNodes, numOfShards)
			cfg.SystemSCConfig.ESDTSystemSCConfig.ESDTPrefix = "sov1"
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

	// Step 2 - generate issue tx
	initialSupply := big.NewInt(144)
	issuedToken := issueESDT(t, cs, initialAddrBytes, "TKN", initialSupply)
	require.True(t, strings.HasPrefix(issuedToken, "sov1-TKN-"))

	// Step 3 - send issued esdt
	receiver := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	receiverBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(receiver)
	receivedTokens := big.NewInt(11)
	require.Nil(t, err)
	transferESDT(t, cs, initialAddrBytes, receiverBytes, receivedTokens, issuedToken)

	// Step 4 - check balances
	requireAccountHasToken(t, cs, issuedToken, receiver, receivedTokens)
	requireAccountHasToken(t, cs, issuedToken, initialAddress, big.NewInt(0).Sub(initialSupply, receivedTokens))
}

func issueESDT(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	sender []byte,
	tokenName string,
	supply *big.Int,
) string {
	// Step 2 - generate issue tx
	issueValue, _ := big.NewInt(0).SetString(issuePrice, 10)
	dataField := fmt.Sprintf("issue@%s@%s@%s@01@63616e55706772616465@74727565@63616e57697065@74727565@63616e467265657a65@74727565",
		hex.EncodeToString([]byte(tokenName)), hex.EncodeToString([]byte(tokenName)), hex.EncodeToString(supply.Bytes()))
	tx := &transaction.Transaction{
		Nonce:     0,
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

	issuedTokens, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs("FungibleESDT")
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(issuedTokens), 1)

	for _, issuedToken := range issuedTokens {
		if strings.Contains(issuedToken, tokenName) {
			log.Info("issued token", "token", issuedToken)
			return issuedToken
		}
	}

	require.Fail(t, "could not create token")
	return ""
}

func transferESDT(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	sender, receiver []byte,
	value *big.Int,
	tokenName string,
) {
	tokenIDHexEncoded := hex.EncodeToString([]byte(tokenName))
	valueHexEncoded := hex.EncodeToString(value.Bytes())
	tx := &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(0),
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(fmt.Sprintf("ESDTTransfer@%s@%s", tokenIDHexEncoded, valueHexEncoded)),
		GasLimit:  100_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	sendTxs, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, sendTxs)
}

func requireAccountHasToken(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	token string,
	address string,
	value *big.Int,
) {
	tokens, _, err := cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens(address, dataApi.AccountQueryOptions{})
	require.Nil(t, err)

	tokenData, found := tokens[token]
	require.True(t, found)
	require.Equal(t, tokenData, &esdt.ESDigitalToken{Value: value})
}
