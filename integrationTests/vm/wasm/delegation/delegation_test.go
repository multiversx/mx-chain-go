//go:build !race

package delegation

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process/factory"
	systemVm "github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

var NewBalance = wasm.NewBalance
var NewBalanceBig = wasm.NewBalanceBig
var RequireAlmostEquals = wasm.RequireAlmostEquals

func TestDelegation_Claims(t *testing.T) {
	// TODO reinstate test after Wasm VM pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	context := wasm.SetupTestContext(t)
	defer context.Close()

	// Genesis
	deployDelegation(context)
	addNodes(context, 2500, 2)

	err := context.ExecuteSC(&context.Alice, "stakeGenesis@"+NewBalance(3000).ToHex())
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Bob, "stakeGenesis@00"+NewBalance(2000).ToHex())
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "activateGenesis")
	require.Nil(t, err)

	require.Equal(t, 3, int(context.QuerySCInt("getNumUsers", [][]byte{})))
	require.Equal(t, NewBalance(3000).Value, context.QuerySCBigInt("getUserStake", [][]byte{context.Alice.Address}))
	require.Equal(t, NewBalance(2000).Value, context.QuerySCBigInt("getUserStake", [][]byte{context.Bob.Address}))

	// No rewards yet
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, NewBalance(5000).Value, context.QuerySCBigInt("getTotalActiveStake", [][]byte{}))

	// Blockchain continues
	context.GoToEpoch(1)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(500).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(500).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, NewBalance(300).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, NewBalance(200).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))

	context.GoToEpoch(2)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(500).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(1000).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, NewBalance(600).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, NewBalance(400).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))

	// Alice, Bob and Carol claim their rewards
	context.TakeAccountBalanceSnapshot(&context.Alice)
	context.TakeAccountBalanceSnapshot(&context.Bob)
	context.TakeAccountBalanceSnapshot(&context.Carol)

	context.GasLimit = 30000000
	err = context.ExecuteSC(&context.Alice, "claimRewards")
	require.Nil(t, err)
	require.Equal(t, 8148760, int(context.LastConsumedFee))
	RequireAlmostEquals(t, NewBalance(600), NewBalanceBig(context.GetAccountBalanceDelta(&context.Alice)))

	err = context.ExecuteSC(&context.Bob, "claimRewards")
	require.Nil(t, err)
	require.Equal(t, 8059660, int(context.LastConsumedFee))
	RequireAlmostEquals(t, NewBalance(400), NewBalanceBig(context.GetAccountBalanceDelta(&context.Bob)))

	err = context.ExecuteSC(&context.Carol, "claimRewards")
	require.Equal(t, errors.New("unknown caller"), err)
}

func TestDelegation_WithManyUsers_Claims(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var err error

	stakePerNode := 2500
	numNodes := 1000
	totalStaked := stakePerNode * numNodes
	numUsers := 100
	stakePerUser := totalStaked / numUsers

	gasSchedule, err := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	require.Nil(t, err)
	gasSchedule["MaxPerTransaction"]["MaxNumberOfTrieReadsPerTx"] = 100

	context := wasm.SetupTestContextWithGasSchedule(t, gasSchedule)
	defer context.Close()

	context.InitAdditionalParticipants(numUsers)

	// Genesis
	deployDelegation(context)
	addNodes(context, stakePerNode, numNodes)

	for _, user := range context.Participants {
		err = context.ExecuteSC(user, "stakeGenesis@"+NewBalance(stakePerUser).ToHex())
		require.Nil(t, err)
	}

	err = context.ExecuteSC(&context.Owner, "activateGenesis")
	require.Nil(t, err)
	require.Equal(t, numUsers+1, int(context.QuerySCInt("getNumUsers", [][]byte{})))

	// Blockchain continues
	context.GoToEpoch(1)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(5000).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(5000).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))

	context.GoToEpoch(2)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(5000).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(10000).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))

	// All users claim their rewards
	for _, user := range context.Participants {
		context.TakeAccountBalanceSnapshot(user)

		context.GasLimit = 30000000
		err = context.ExecuteSC(user, "claimRewards")
		require.Nil(t, err)
		require.LessOrEqual(t, int(context.LastConsumedFee), 25000000)
		RequireAlmostEquals(t, NewBalance(10000/numUsers), NewBalanceBig(context.GetAccountBalanceDelta(user)))
	}
}

func deployDelegation(context *wasm.TestContext) {
	delegationWasmPath := "../testdata/delegation/delegation.wasm"
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@03E8@00@030D40@030D40"

	err := context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Nil(context.T, err)
}

func addNodes(context *wasm.TestContext, stakePerNode int, numNodes int) {
	err := context.ExecuteSC(&context.Owner, "setStakePerNode@"+NewBalance(stakePerNode).ToHex())
	require.Nil(context.T, err)

	addNodesArguments := make([]string, 0, numNodes*2)
	for tag := 0; tag < numNodes; tag++ {
		tagBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(tagBytes, uint32(tag))

		blsKey := bytes.Repeat(tagBytes, 96/len(tagBytes))
		blsSignature := bytes.Repeat(tagBytes, 32/len(tagBytes))

		addNodesArguments = append(addNodesArguments, hex.EncodeToString(blsKey))
		addNodesArguments = append(addNodesArguments, hex.EncodeToString(blsSignature))
	}

	err = context.ExecuteSC(&context.Owner, "addNodes@"+strings.Join(addNodesArguments, "@"))
	require.Nil(context.T, err)
	require.Equal(context.T, numNodes, int(context.QuerySCInt("getNumNodes", [][]byte{})))
	fmt.Println("addNodes consumed (gas):", context.LastConsumedFee)
}

func TestDelegationProcessManyAotInProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	delegationProcessManyTimes(t, "../testdata/delegation/delegation_v0_5_1_full.wasm", 2, 1)
}

func TestDelegationShrinkedProcessManyAotInProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	delegationProcessManyTimes(t, "../testdata/delegation/delegation_v0_5_2_full.wasm", 2, 1)
}

func TestDelegationProcessManyTimeCompileWithOutOfProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	delegationProcessManyTimes(t, "../testdata/delegation/delegation_v0_5_1_full.wasm", 100, 1)
}

func delegationProcessManyTimes(t *testing.T, fileName string, txPerBenchmark int, numRun int) {
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(10000000000000)
	gasPrice := uint64(1)
	gasLimit := uint64(10000000000)

	scCode := wasm.GetSCCode(fileName)
	// 17918321 - stake in active - 11208675 staking in waiting - 28276371 - unstake from active
	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	testContext, err := vm.CreateTxProcessorWasmVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		config.EnableEpochs{},
	)
	require.Nil(t, err)

	defer testContext.Close()

	value := big.NewInt(10)
	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.WasmVirtualMachine)
	serviceFeePer10000 := 3000
	blocksBeforeUnBond := 60

	totalDelegationCap := big.NewInt(0).Mul(big.NewInt(int64(txPerBenchmark)), value)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		wasm.CreateDeployTxData(scCode)+
			"@"+hex.EncodeToString(systemVm.ValidatorSCAddress)+"@"+core.ConvertToEvenHex(serviceFeePer10000)+
			"@"+core.ConvertToEvenHex(serviceFeePer10000)+"@"+core.ConvertToEvenHex(blocksBeforeUnBond)+
			"@"+hex.EncodeToString(value.Bytes())+"@"+hex.EncodeToString(totalDelegationCap.Bytes()),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	ownerNonce++

	testAddresses := createTestAddresses(uint64(txPerBenchmark * 2))
	for _, testAddress := range testAddresses {
		_, _ = vm.CreateAccount(testContext.Accounts, testAddress, 0, big.NewInt(10000000000000))
	}
	_, _ = testContext.Accounts.Commit()

	for j := 0; j < numRun; j++ {
		start := time.Now()
		for i := 0; i < txPerBenchmark; i++ {
			testAddress := testAddresses[i]
			nonce := uint64(j * 2)
			tx = &transactionData.Transaction{
				Nonce:    nonce,
				Value:    big.NewInt(0).Set(value),
				SndAddr:  testAddress,
				RcvAddr:  scAddress,
				Data:     []byte("stake"),
				GasPrice: gasPrice,
				GasLimit: gasLimit,
			}

			returnCodeAfterProcess, _ := testContext.TxProcessor.ProcessTransaction(tx)
			if returnCodeAfterProcess != vmcommon.Ok {
				fmt.Printf("return code %s \n", returnCodeAfterProcess.String())
			}
		}

		elapsedTime := time.Since(start)
		fmt.Printf("time elapsed to process %d stake on delegation %s \n", txPerBenchmark, elapsedTime.String())
		printGasConsumed(testContext, "stake to active", gasLimit)

		start = time.Now()
		for i := txPerBenchmark; i < txPerBenchmark*2; i++ {
			testAddress := testAddresses[i]
			nonce := uint64(j)
			tx = &transactionData.Transaction{
				Nonce:    nonce,
				Value:    big.NewInt(0).Set(value),
				SndAddr:  testAddress,
				RcvAddr:  scAddress,
				Data:     []byte("stake"),
				GasPrice: gasPrice,
				GasLimit: gasLimit,
			}

			returnCodeAfterProcess, _ := testContext.TxProcessor.ProcessTransaction(tx)
			if returnCodeAfterProcess != vmcommon.Ok {
				fmt.Printf("return code %s \n", returnCodeAfterProcess.String())
			}
		}

		elapsedTime = time.Since(start)
		fmt.Printf("time elapsed to process %d stake to waiting list %s \n", txPerBenchmark, elapsedTime.String())
		printGasConsumed(testContext, "stake to waiting", gasLimit)

		start = time.Now()
		for i := 0; i < txPerBenchmark; i++ {
			testAddress := testAddresses[i]
			tx = &transactionData.Transaction{
				Nonce:    uint64(j*2 + 1),
				Value:    big.NewInt(0),
				SndAddr:  testAddress,
				RcvAddr:  scAddress,
				Data:     []byte("unStake@" + hex.EncodeToString(value.Bytes())),
				GasPrice: gasPrice,
				GasLimit: gasLimit,
			}

			returnCodeAfterProcess, _ := testContext.TxProcessor.ProcessTransaction(tx)
			if returnCodeAfterProcess != vmcommon.Ok {
				fmt.Printf("return code %s \n", returnCodeAfterProcess.String())
			}
		}

		elapsedTime = time.Since(start)
		fmt.Printf("time elapsed to process %d unStake on delegation %s \n", txPerBenchmark, elapsedTime.String())
		_, _ = testContext.Accounts.Commit()
		printGasConsumed(testContext, "unStake from waiting", gasLimit)

		start = time.Now()
		tx = &transactionData.Transaction{
			Nonce:    ownerNonce,
			Value:    big.NewInt(0),
			SndAddr:  ownerAddressBytes,
			RcvAddr:  scAddress,
			Data:     []byte("getFullWaitingList"),
			GasPrice: gasPrice,
			GasLimit: gasLimit,
		}

		ownerNonce++
		returnCodeAfterProcess, _ := testContext.TxProcessor.ProcessTransaction(tx)
		if returnCodeAfterProcess != vmcommon.Ok {
			fmt.Printf("return code %s \n", returnCodeAfterProcess.String())
		}

		elapsedTime = time.Since(start)
		fmt.Printf("time elapsed to process getFullWaitingList %s \n", elapsedTime.String())
		_, _ = testContext.Accounts.Commit()

		printGasConsumed(testContext, "getFullWaitingList", gasLimit)
	}
}

func printGasConsumed(testContext *vm.VMTestContext, functionName string, gasLimit uint64) {
	gasRemaining := testContext.GetGasRemaining()
	fmt.Printf("%s was executed, consumed %d gas\n", functionName, gasLimit-gasRemaining)
}

func createTestAddresses(numAddresses uint64) [][]byte {
	testAccounts := make([][]byte, numAddresses)

	for i := uint64(0); i < numAddresses; i++ {
		acc := generateRandomByteArray(32)
		testAccounts[i] = append(testAccounts[i], acc...)
	}

	return testAccounts
}

func generateRandomByteArray(size int) []byte {
	r := make([]byte, size)
	_, _ = rand.Read(r)
	return r
}
