package delegation

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	systemVm "github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationtests/vm/arwen/delegation")

func TestSimulateExecutionOfStakeTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cacheConfig := storageUnit.CacheConfig{
		Name:        "trie",
		Type:        "SizeLRU",
		SizeInBytes: 314572800, //300MB
		Capacity:    500000,
	}
	trieCache, err := storageUnit.NewCache(cacheConfig)
	require.Nil(t, err)
	dbConfig := config.DBConfig{
		FilePath:          "trie",
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}
	persisterFactory := factory.NewPersisterFactory(dbConfig)
	tempDir, err := ioutil.TempDir("", "integrationTest")
	require.Nil(t, err)
	triePersister, err := persisterFactory.Create(tempDir)
	require.Nil(t, err)
	trieStorage, err := storageUnit.NewStorageUnit(trieCache, triePersister)
	require.Nil(t, err)

	defer func() {
		err = trieStorage.DestroyUnit()
		log.LogIfError(err)
	}()

	gasMap, err := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	require.Nil(t, err)

	node := integrationTests.NewTestProcessorNodeWithStorageTrieAndGasModel(
		1,
		0,
		0,
		"",
		trieStorage,
		gasMap,
	)

	totalSupply, _ := big.NewInt(0).SetString("20000000000000000000000000", 10) //20MIL eGLD
	nodeInitialBalance := big.NewInt(0).Set(totalSupply)
	nodeInitialBalance.Div(nodeInitialBalance, big.NewInt(2))
	node.EconomicsData.SetMaxGasLimitPerBlock(1500000000)
	node.EconomicsData.SetMinGasLimit(50000)
	node.EconomicsData.SetMinGasPrice(1000000000)
	node.EconomicsData.SetTotalSupply(totalSupply)
	integrationTests.MintAllNodes([]*integrationTests.TestProcessorNode{node}, nodeInitialBalance)

	numAccounts := 100000
	accountsInitialBalance, _ := big.NewInt(0).SetString("1000000000000000000000", 10) //1000eGLD
	addresses := generateAndMintAccounts(t, node, accountsInitialBalance, numAccounts)

	delegationAddr, err := node.BlockchainHook.NewAddress(node.OwnAccount.Address, node.OwnAccount.Nonce, []byte{5, 0})
	log.Info("delegation contract", "address", integrationTests.TestAddressPubkeyConverter.Encode(delegationAddr))

	deployDelegationSC(t, node)

	_, err = node.AccntState.Commit()
	require.Nil(t, err)

	for j := 0; j < 2; j++ {
		numBatches := 10
		log.Info("starting staking round", "round", j)
		for i := 0; i < numBatches; i++ {
			batch := addresses[i*100 : (i+1)*100]

			sw := core.NewStopWatch()
			sw.Start("do stake")
			doStake(t, node, batch, delegationAddr)
			sw.Stop("do stake")

			logInfo := sw.GetMeasurements()
			logInfo = append(logInfo, "num addresses in batch")
			logInfo = append(logInfo, len(batch))

			log.Info("process took", logInfo...)

			_, err = node.AccntState.Commit()
			require.Nil(t, err)

		}
	}
}

func deployDelegationSC(t *testing.T, node *integrationTests.TestProcessorNode) {
	serviceFeePer10000 := 3000
	blocksBeforeUnBond := 60
	value := big.NewInt(10)

	contractBytes, err := ioutil.ReadFile("../testdata/delegation/delegation_v0_5_2_full.wasm")
	require.Nil(t, err)

	tx := vm.CreateDeployTx(
		node.OwnAccount.Address,
		node.OwnAccount.Nonce,
		big.NewInt(0),
		node.EconomicsData.MinGasPrice(),
		node.EconomicsData.GetMinGasLimit()+uint64(100000000),
		arwen.CreateDeployTxData(hex.EncodeToString(contractBytes))+
			"@"+hex.EncodeToString(systemVm.ValidatorSCAddress)+"@"+core.ConvertToEvenHex(serviceFeePer10000)+
			"@"+core.ConvertToEvenHex(serviceFeePer10000)+"@"+core.ConvertToEvenHex(blocksBeforeUnBond)+
			"@"+hex.EncodeToString(value.Bytes())+"@"+hex.EncodeToString(node.EconomicsData.GenesisTotalSupply().Bytes()),
	)

	retCode, err := node.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, retCode)
}

func generateAndMintAccounts(t *testing.T, node *integrationTests.TestProcessorNode, initialBalance *big.Int, numAccounts int) [][]byte {
	addresses := make([][]byte, 0, numAccounts)

	for i := 0; i < numAccounts; i++ {
		addr := make([]byte, integrationTests.TestAddressPubkeyConverter.Len())
		_, _ = rand.Read(addr)

		account, err := node.AccntState.LoadAccount(addr)
		require.Nil(t, err)

		userAccount := account.(state.UserAccountHandler)
		err = userAccount.AddToBalance(initialBalance)
		require.Nil(t, err)

		err = node.AccntState.SaveAccount(userAccount)
		require.Nil(t, err)

		addresses = append(addresses, addr)
	}

	return addresses
}

func doStake(t *testing.T, node *integrationTests.TestProcessorNode, addresses [][]byte, scAddress []byte) {
	stakeVal, _ := big.NewInt(0).SetString("10000000000000000000", 10) //10eGLD

	for _, addr := range addresses {
		doStakeOneAddress(t, node, addr, stakeVal, scAddress)
	}
}

func doStakeOneAddress(t *testing.T, node *integrationTests.TestProcessorNode, address []byte, stakeVal *big.Int, scAddress []byte) {
	accnt, _ := node.AccntState.GetExistingAccount(address)

	tx := vm.CreateTx(
		t,
		address,
		scAddress,
		accnt.GetNonce(),
		stakeVal,
		node.EconomicsData.MinGasPrice(),
		node.EconomicsData.MinGasLimit()+100000000,
		"stake",
	)

	retCode, err := node.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, retCode)
}
