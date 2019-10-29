package systemVM

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	"github.com/stretchr/testify/assert"
)

func TestRunWithTransferAndGasShouldRunSCCode(t *testing.T) {
	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3
	firstSkInShard := uint32(0)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	var metachainNodes []*integrationTests.TestProcessorNode
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == sharding.MetachainShardId {
			metachainNodes = append(metachainNodes, node)
		}
	}
	assert.NotNil(t, metachainNodes)

	nodes[0] = integrationTests.NewTestProcessorNode(uint32(numOfShards), 0, firstSkInShard, integrationTests.GetConnectableAddress(advertiser))
	integrationTests.CreateAccountForNodes(nodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	scAddress := factory.StakingSCAddress

	initialBalance, _ := big.NewInt(0).SetString("5000000000000000000000000", 10)
	senderBytes, _ := hex.DecodeString("53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8")
	integrationTests.MintAddress(nodes[0].AccntState, senderBytes, big.NewInt(0).Set(initialBalance))

	///////////------- send a stake tx and check sender's balance
	txValue, _ := big.NewInt(0).SetString("500000000000000000000000", 10)
	stakeTx := &transaction.Transaction{
		Nonce:    0,
		Value:    txValue,
		SndAddr:  senderBytes,
		RcvAddr:  scAddress,
		Data:     "stake",
		GasPrice: integrationTests.MinTxGasPrice,
		GasLimit: integrationTests.MinTxGasLimit * 5,
	}

	err := nodes[0].TxProcessor.ProcessTransaction(stakeTx, 0)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	sndAcc := getAccountFromAddrBytes(nodes[0].AccntState, senderBytes)
	assert.Equal(t, big.NewInt(0).Sub(initialBalance, txValue), sndAcc.Balance)

	time.Sleep(5 * time.Second)

	/////////------ send another stake tx. it shouldn't be executed
	stakeTx = &transaction.Transaction{
		Nonce:    1,
		Value:    txValue,
		SndAddr:  senderBytes,
		RcvAddr:  scAddress,
		Data:     "stake",
		GasPrice: integrationTests.MinTxGasPrice,
		GasLimit: integrationTests.MinTxGasLimit * 5,
	}

	err = nodes[0].TxProcessor.ProcessTransaction(stakeTx, 1)
	assert.Nil(t, err)

	time.Sleep(time.Second)
	sndAcc = getAccountFromAddrBytes(nodes[0].AccntState, senderBytes)
	// TODO : de-comment these assert when the entire functionality is done
	//assert.Equal(t, big.NewInt(0).Sub(initialBalance, txValue), sndAcc.Balance)

	/////////------ send an unStake tx
	unStakeTx := &transaction.Transaction{
		Nonce:    2,
		Value:    big.NewInt(0),
		SndAddr:  senderBytes,
		RcvAddr:  scAddress,
		Data:     "unStake",
		GasPrice: integrationTests.MinTxGasPrice,
		GasLimit: integrationTests.MinTxGasLimit * 5,
	}

	err = nodes[0].TxProcessor.ProcessTransaction(unStakeTx, 2)
	assert.Nil(t, err)

	time.Sleep(time.Second)
	sndAcc = getAccountFromAddrBytes(nodes[0].AccntState, senderBytes)
	//assert.Equal(t, initialBalance, sndAcc.Balance)

	////////----- send an finalizeUnstake

	time.Sleep(5 * time.Second)
	finalizeUnStakeTx := &transaction.Transaction{
		Nonce:    3,
		Value:    big.NewInt(0),
		SndAddr:  senderBytes,
		RcvAddr:  scAddress,
		Data:     "finalizeUnStake",
		GasPrice: integrationTests.MinTxGasPrice,
		GasLimit: integrationTests.MinTxGasLimit * 5,
	}

	err = nodes[0].TxProcessor.ProcessTransaction(finalizeUnStakeTx, 2)
	assert.Nil(t, err)

	time.Sleep(time.Second)
	sndAcc = getAccountFromAddrBytes(nodes[0].AccntState, senderBytes)
	//assert.Equal(t, initialBalance, sndAcc.Balance)
}

func getAccountFromAddrBytes(accState state.AccountsAdapter, address []byte) *state.Account {
	addrCont, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(address)
	sndrAcc, _ := accState.GetExistingAccount(addrCont)

	sndAccSt, _ := sndrAcc.(*state.Account)

	return sndAccSt
}
