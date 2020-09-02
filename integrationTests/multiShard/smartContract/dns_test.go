package smartContract

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func TestSCCallingDNSUserNames(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, players, idxProposers, advertiser := prepareNodesAndPlayers()

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	dnsRegisterValue, sortedDNSAddresses := getDNSContractsData(nodes[0])

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	userNames := sendRegisterUserNameTxForPlayers(players, nodes, sortedDNSAddresses, dnsRegisterValue)

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 15
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	checkUserNamesAreSetCorrectly(t, players, nodes, userNames)
}

func TestSCCallingDNSUserNamesTwice(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, players, idxProposers, advertiser := prepareNodesAndPlayers()

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	dnsRegisterValue, sortedDNSAddresses := getDNSContractsData(nodes[0])

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	userNames := sendRegisterUserNameTxForPlayers(players, nodes, sortedDNSAddresses, dnsRegisterValue)

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	_ = sendRegisterUserNameTxForPlayers(players, nodes, sortedDNSAddresses, dnsRegisterValue)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	checkUserNamesAreSetCorrectly(t, players, nodes, userNames)
}

func prepareNodesAndPlayers() ([]*integrationTests.TestProcessorNode, []*integrationTests.TestWalletAccount, []int, p2p.Messenger) {
	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	genesisFile := "testdata/smartcontracts.json"
	nodes := integrationTests.CreateNodesWithFullGenesis(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
		genesisFile,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	numPlayers := 6
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, 0)
	}

	initialVal := big.NewInt(10000000000000)
	initialVal.Mul(initialVal, initialVal)
	fmt.Printf("Initial minted sum: %s\n", initialVal.String())
	integrationTests.MintAllNodes(nodes, initialVal)
	integrationTests.MintAllPlayers(nodes, players, initialVal)

	return nodes, players, idxProposers, advertiser
}

func getDNSContractsData(node *integrationTests.TestProcessorNode) (*big.Int, []string) {
	dnsRegisterValue := big.NewInt(100)
	genesisSCs := node.SmartContractParser.InitialSmartContracts()
	for _, genesisSC := range genesisSCs {
		if genesisSC.GetType() == genesis.DNSType {
			decodedValue, _ := hex.DecodeString(genesisSC.GetInitParameters())
			dnsRegisterValue.SetBytes(decodedValue)
			break
		}
	}

	mapDNSAddresses, _ := node.SmartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	sortedDNSAddresses := make([]string, 0, len(mapDNSAddresses))
	for address := range mapDNSAddresses {
		sortedDNSAddresses = append(sortedDNSAddresses, address)
	}
	sort.Slice(sortedDNSAddresses, func(i, j int) bool {
		return sortedDNSAddresses[i][31] < sortedDNSAddresses[j][31]
	})

	return dnsRegisterValue, sortedDNSAddresses
}

func sendRegisterUserNameTxForPlayers(
	players []*integrationTests.TestWalletAccount,
	nodes []*integrationTests.TestProcessorNode,
	sortedDNSAddresses []string,
	dnsRegisterValue *big.Int,
) []string {

	gasLimit := uint64(200000)
	userNames := make([]string, len(players))
	for i, player := range players {
		userName := generateNewUserName()
		scAddress := selectDNSAddressFromUserName(sortedDNSAddresses, userName)
		integrationTests.PlayerSendsTransaction(
			nodes,
			player,
			[]byte(scAddress),
			dnsRegisterValue,
			"register@"+hex.EncodeToString([]byte(userName)),
			gasLimit,
		)
		userNames[i] = userName
	}
	return userNames
}

func checkUserNamesAreSetCorrectly(
	t *testing.T,
	players []*integrationTests.TestWalletAccount,
	nodes []*integrationTests.TestProcessorNode,
	userNames []string,
) {
	for i, player := range players {
		playerShID := nodes[0].ShardCoordinator.ComputeId(player.Address)
		for _, node := range nodes {
			if node.ShardCoordinator.SelfId() != playerShID {
				continue
			}

			acnt, _ := node.AccntState.GetExistingAccount(player.Address)
			userAcc, _ := acnt.(state.UserAccountHandler)

			hashedUserName := keccak.Keccak{}.Compute(userNames[i])
			assert.Equal(t, hashedUserName, userAcc.GetUserName())
		}
	}
}

func selectDNSAddressFromUserName(sortedDNSAddresses []string, userName string) string {
	hashedAddr := keccak.Keccak{}.Compute(userName)
	return sortedDNSAddresses[hashedAddr[31]]
}

func generateNewUserName() string {
	buff := make([]byte, 10)
	_, _ = rand.Read(buff)
	return hex.EncodeToString(buff)
}
