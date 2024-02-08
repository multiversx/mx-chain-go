package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"

	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestChainSimulator_AddANewValidatorsAfterStakingV4(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	_ = logger.SetLogLevel("*:NONE")
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	cm, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       15,
		MetaChainMinNodes:      15,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			cfg.GeneralConfig.ValidatorStatistics.CacheRefreshIntervalInSec = 1
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cm)

	//Wait for staking v4 to activate
	err = cm.GenerateBlocks(150)
	require.Nil(t, err)

	// Step 1 --- add 20 validator keys in the chain simulator
	numOfNodes := 20
	validatorSecretKeysBytes, blsKeys, _ := chainSimulator.GenerateBlsPrivateKeys(numOfNodes)
	err = cm.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions - 1 000 000
	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "1000000000000000000000000",
		},
	})
	require.Nil(t, err)

	// Step 3 --- generate and send a stake transaction with the BLS keys of the validators key that were added at step 1
	validatorData := ""
	for _, blsKey := range blsKeys {
		validatorData += fmt.Sprintf("@%s@010101", blsKey)
	}

	numOfNodesHex := hex.EncodeToString(big.NewInt(int64(numOfNodes)).Bytes())
	stakeValue, _ := big.NewInt(0).SetString("51000000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     stakeValue,
		SndAddr:   newValidatorOwnerBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("stake@%s%s", numOfNodesHex, validatorData)),
		GasLimit:  500_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txFromNetwork, _ := cm.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.NotNil(t, txFromNetwork)

	err = cm.GenerateBlocks(1)
	require.Nil(t, err)

	auctionListR, err := cm.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().AuctionListApi()

	topup := auctionListR[0].TopUpPerNode
	auctionList := auctionListR[0].AuctionList

	fmt.Println("topuppernode", topup)
	fmt.Println("qualifiedtopup", auctionListR[0].QualifiedTopUp)
	// print the size of the auction list slice
	fmt.Println("auctionList", len(auctionList))
	// print how many of them are qualified
	qualified := 0
	for _, node := range auctionList {
		if node.Qualified {
			qualified++
		}
	}
	fmt.Println("qualified", qualified)

	require.Nil(t, err)

	err = cm.GenerateBlocks(100)
	require.Nil(t, err)
}

func TestChainSimulator_DelegationManagerScen9(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:NONE")

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	cm, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       100,
		MetaChainMinNodes:      100,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			cfg.GeneralConfig.ValidatorStatistics.CacheRefreshIntervalInSec = 1
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cm)

	//delegation contract address
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6"
	rcvAddrBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	//address A with 10 000 EGLD
	delegationManagerA := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	delegationManagerABytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(delegationManagerA)

	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "1000000000000000000000000",
		},
	})
	require.Nil(t, err)

	//address B with 10 000 EGLD
	delegationManagerB := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	delegationManagerBBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(delegationManagerB)

	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx",
			Balance: "1000000000000000000000000",
		},
	})
	require.Nil(t, err)

	//delegatorA with 2 000 EGLD
	//delegatorA := "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8"
	//delegatorABytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(delegatorA)

	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
			Balance: "200000000000000000000000",
		},
	})
	require.Nil(t, err)

	//delegatorB with 2 000 EGLD
	//delegatorB := "erd1kyaqzaprcdnv4luvanah0gfxzzsnpaygsy6pytrexll2urtd05ts9vegu7"
	//delegatorBBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(delegatorB)

	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1kyaqzaprcdnv4luvanah0gfxzzsnpaygsy6pytrexll2urtd05ts9vegu7",
			Balance: "200000000000000000000000",
		},
	})
	require.Nil(t, err)

	// //Wait for one epoch
	err = cm.GenerateBlocks(30)
	require.Nil(t, err)

	// Step 1 --- With delegationManagerA createNewDelegationContract 1250EGLD
	delegationValue, _ := big.NewInt(0).SetString("1250000000000000000000", 10)
	denominatedDelCap, _ := big.NewInt(0).SetString("51000000000000000000000", 10)
	denominatedDelCapHex := hex.EncodeToString(denominatedDelCap.Bytes())
	serviceFee := []byte{100}
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     delegationValue,
		SndAddr:   delegationManagerABytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("createNewDelegationContract@%s@%s", denominatedDelCapHex, hex.EncodeToString(serviceFee))),
		GasLimit:  500_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txFromNetwork, err := cm.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txFromNetwork)
	data := txFromNetwork.SmartContractResults[0].Data
	parts := strings.Split(data, "@")
	require.Equal(t, 3, len(parts))

	require.Equal(t, hex.EncodeToString([]byte("ok")), parts[1])
	delegationContractAddressHex, _ := hex.DecodeString(parts[2])
	delegationContractAddress, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(delegationContractAddressHex)
	//print delegationContractAddress
	fmt.Println("delegationContractAddress", delegationContractAddress)
	err = cm.GenerateBlocks(1)
	require.Nil(t, err)

	// Step 2 --- With delegationManagerB createNewDelegationContract 1250EGLD

	tx = &transaction.Transaction{
		Nonce:     0,
		Value:     delegationValue,
		SndAddr:   delegationManagerBBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("createNewDelegationContract@%s@%s", denominatedDelCapHex, hex.EncodeToString(serviceFee))),
		GasLimit:  500_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txFromNetwork, err = cm.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txFromNetwork)
	data = txFromNetwork.SmartContractResults[0].Data
	parts = strings.Split(data, "@")
	require.Equal(t, 3, len(parts))

	require.Equal(t, hex.EncodeToString([]byte("ok")), parts[1])
	delegationContractAddressBHex, _ := hex.DecodeString(parts[2])
	delegationContractAddressB, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(delegationContractAddressBHex)
	//print delegationContractAddress
	fmt.Println("delegationContractAddress", delegationContractAddressB)
	err = cm.GenerateBlocks(1)
	require.Nil(t, err)

	err = cm.GenerateBlocks(1)
	require.Nil(t, err)
}
