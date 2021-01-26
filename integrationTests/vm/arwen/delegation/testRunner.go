package delegation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	systemVm "github.com/ElrondNetwork/elrond-go/vm"
)

var log = logger.GetOrCreate("integrationtests/vm/arwen/delegation")

// RunDelegationStressTest will call the stake function and measure the time duration for each call
func RunDelegationStressTest(
	delegationFilename string,
	numRuns uint32,
	numBatches uint32,
	numTxPerBatch uint32,
	numQueriesPerBatch uint32,
	gasSchedule map[string]map[string]uint64,
) ([]time.Duration, error) {

	cacheConfig := storageUnit.CacheConfig{
		Name:        "trie",
		Type:        "SizeLRU",
		SizeInBytes: 314572800, //300MB
		Capacity:    500000,
	}
	trieCache, err := storageUnit.NewCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	dbConfig := config.DBConfig{
		FilePath:          "trie",
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}
	persisterFactory := factory.NewPersisterFactory(dbConfig)
	tempDir, err := ioutil.TempDir("", "integrationTest")
	if err != nil {
		return nil, err
	}

	triePersister, err := persisterFactory.Create(tempDir)
	if err != nil {
		return nil, err
	}

	trieStorage, err := storageUnit.NewStorageUnit(trieCache, triePersister)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = trieStorage.DestroyUnit()
		log.LogIfError(err)
	}()

	node := integrationTests.NewTestProcessorNodeWithStorageTrieAndGasModel(
		1,
		0,
		0,
		"",
		trieStorage,
		gasSchedule,
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
	addresses, err := generateAndMintAccounts(node, accountsInitialBalance, numAccounts)
	if err != nil {
		return nil, err
	}

	delegationAddr, err := node.BlockchainHook.NewAddress(node.OwnAccount.Address, node.OwnAccount.Nonce, []byte{5, 0})
	log.Debug("delegation contract", "address", integrationTests.TestAddressPubkeyConverter.Encode(delegationAddr))

	err = deployDelegationSC(node, delegationFilename)
	if err != nil {
		return nil, err
	}

	_, err = node.AccntState.Commit()
	if err != nil {
		return nil, err
	}

	stopRequest := make(chan struct{}, 1)
	wg := sync.WaitGroup{}
	mutExecutionError := sync.Mutex{}
	var executionError error
	if numQueriesPerBatch > 0 {
		wg.Add(1)
		copiedAddresses := make([][]byte, 0)
		for _, address := range addresses {
			copiedAddresses = append(copiedAddresses, address)
		}

		scQuery := node.SCQueryService
		go func() {
			defer wg.Done()

			getClaimableRewards := &process.SCQuery{
				ScAddress:  delegationAddr,
				FuncName:   "getClaimableRewards",
				CallerAddr: delegationAddr,
				CallValue:  big.NewInt(0),
				Arguments:  [][]byte{},
			}

			getUserStakeByType := &process.SCQuery{
				ScAddress:  delegationAddr,
				FuncName:   "getUserStakeByType",
				CallerAddr: delegationAddr,
				CallValue:  big.NewInt(0),
				Arguments:  [][]byte{},
			}

			for j := uint32(0); j < numBatches; j++ {
				for i := uint32(0); i < numQueriesPerBatch; i++ {
					getClaimableRewards.Arguments = [][]byte{copiedAddresses[j]}
					getUserStakeByType.Arguments = [][]byte{copiedAddresses[j]}

					_, localErrQuery := scQuery.ExecuteQuery(getClaimableRewards)
					if localErrQuery != nil {
						mutExecutionError.Lock()
						executionError = localErrQuery
						mutExecutionError.Unlock()
					}

					_, localErrQuery = scQuery.ExecuteQuery(getUserStakeByType)
					if localErrQuery != nil {
						mutExecutionError.Lock()
						executionError = localErrQuery
						mutExecutionError.Unlock()
					}
				}

				select {
				case <-time.After(time.Second):
				case <-stopRequest:
					return
				}
			}
		}()
	}

	benchmarks := make([]time.Duration, 0)
	for j := uint32(0); j < numRuns; j++ {
		log.Info("starting staking round", "round", j)
		for i := uint32(0); i < numBatches; i++ {
			batch := addresses[i*numTxPerBatch : (i+1)*numTxPerBatch]

			startTime := time.Now()
			err = doStake(node, batch, delegationAddr)
			benchmarks = append(benchmarks, time.Since(startTime))
			if err != nil {
				mutExecutionError.Lock()
				executionError = err
				mutExecutionError.Unlock()
			}

			_, err = node.AccntState.Commit()
			if err != nil {
				mutExecutionError.Lock()
				executionError = err
				mutExecutionError.Unlock()
			}
		}
	}

	stopRequest <- struct{}{}
	wg.Wait()

	return benchmarks, executionError
}

func deployDelegationSC(node *integrationTests.TestProcessorNode, delegationFilename string) error {
	serviceFeePer10000 := 3000
	blocksBeforeUnBond := 60
	value := big.NewInt(10)

	contractBytes, err := ioutil.ReadFile(delegationFilename)
	if err != nil {
		return err
	}

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
	if err != nil {
		return err
	}
	if retCode != vmcommon.Ok {
		return fmt.Errorf("return code on deploy is not Ok, but %s", retCode.String())
	}

	return nil
}

func generateAndMintAccounts(node *integrationTests.TestProcessorNode, initialBalance *big.Int, numAccounts int) ([][]byte, error) {
	addresses := make([][]byte, 0, numAccounts)

	for i := 0; i < numAccounts; i++ {
		addr := make([]byte, integrationTests.TestAddressPubkeyConverter.Len())
		_, _ = rand.Read(addr)

		account, err := node.AccntState.LoadAccount(addr)
		if err != nil {
			return nil, err
		}

		userAccount := account.(state.UserAccountHandler)
		err = userAccount.AddToBalance(initialBalance)
		if err != nil {
			return nil, err
		}

		err = node.AccntState.SaveAccount(userAccount)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}

func doStake(node *integrationTests.TestProcessorNode, addresses [][]byte, scAddress []byte) error {
	stakeVal, _ := big.NewInt(0).SetString("10000000000000000000", 10) //10eGLD

	for _, addr := range addresses {
		err := doStakeOneAddress(node, addr, stakeVal, scAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

func doStakeOneAddress(node *integrationTests.TestProcessorNode, address []byte, stakeVal *big.Int, scAddress []byte) error {
	accnt, _ := node.AccntState.GetExistingAccount(address)

	tx := vm.CreateTx(
		address,
		scAddress,
		accnt.GetNonce(),
		stakeVal,
		node.EconomicsData.MinGasPrice(),
		node.EconomicsData.MinGasLimit()+100000000,
		"stake",
	)

	retCode, err := node.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return err
	}
	if retCode != vmcommon.Ok {
		return fmt.Errorf("return code on stake execution is not Ok, but %s", retCode.String())
	}

	return nil
}
