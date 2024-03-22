package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	api2 "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig             = "../../../cmd/node/config/"
	maxNumOfBlockToGenerateWhenExecutingTx = 7
)

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
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = "5000000000000000000"
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "10000000000000000000000",
		},
	})
	require.Nil(t, err)

	rcvAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode("erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl")
	//keyValueMap := map[string]string{
	//	"01": "01",
	//	"02": "02",
	//}
	//err = cs.SetKeyValueForAddress("erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t", keyValueMap)
	//require.Nil(t, err)

	// Step 3 --- generate and send a stake transaction with the BLS key of the validator key that was added at step 1
	stakeValue, _ := big.NewInt(0).SetString("5000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     stakeValue,
		SndAddr:   rcvAddrBytes,
		RcvAddr:   vm.ESDTSCAddress,
		Data:      []byte("issue@4141414141@41414141@6f@01@63616e55706772616465@74727565@63616e57697065@74727565@63616e467265657a65@74727565"),
		GasLimit:  100_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	issuedTokens, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs("FungibleESDT")
	require.Nil(t, err)
	require.Len(t, issuedTokens, 1)

	tokenIDHexEncoded := hex.EncodeToString([]byte(issuedTokens[0]))

	rcv2, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode("erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx")
	tx = &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(0),
		SndAddr:   rcvAddrBytes,
		RcvAddr:   rcv2,
		Data:      []byte(fmt.Sprintf("ESDTTransfer@%s@0b", tokenIDHexEncoded)),
		GasLimit:  100_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	tokens, _, err := cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens("erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx", api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotEmpty(t, tokens)
}
