package dns

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/relayedTx"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	nrRoundsToPropagateMultiShard := 25
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	checkUserNamesAreSetCorrectly(t, players, nodes, userNames, sortedDNSAddresses)
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

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	checkUserNamesAreSetCorrectly(t, players, nodes, userNames, sortedDNSAddresses)
}

func TestDNSandRelayedTxNormal(t *testing.T) {
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

	relayer := createAndMintRelayer(nodes)
	dnsRegisterValue, sortedDNSAddresses := getDNSContractsData(nodes[0])

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	userNames := sendRegisterUserNameAsRelayedTx(relayer, players, nodes, sortedDNSAddresses, dnsRegisterValue)

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 30
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	checkUserNamesAreSetCorrectly(t, players, nodes, userNames, sortedDNSAddresses)
}

func createAndMintRelayer(nodes []*integrationTests.TestProcessorNode) *integrationTests.TestWalletAccount {
	relayer := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, 0)

	initialVal := big.NewInt(10000000000000)
	initialVal.Mul(initialVal, initialVal)
	integrationTests.MintAllPlayers(nodes, []*integrationTests.TestWalletAccount{relayer}, initialVal)
	return relayer
}

func prepareNodesAndPlayers() ([]*integrationTests.TestProcessorNode, []*integrationTests.TestWalletAccount, []int, p2p.Messenger) {
	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	genesisFile := "smartcontracts.json"
	nodes := integrationTests.CreateNodesWithFullGenesis(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
		genesisFile,
	)

	for _, node := range nodes {
		node.EconomicsData.SetMaxGasLimitPerBlock(1500000000)
	}

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

func sendRegisterUserNameAsRelayedTx(
	relayer *integrationTests.TestWalletAccount,
	players []*integrationTests.TestWalletAccount,
	nodes []*integrationTests.TestProcessorNode,
	sortedDNSAddresses []string,
	dnsRegisterValue *big.Int,
) []string {

	userNames := make([]string, len(players))
	gasLimit := uint64(2000000)
	for i, player := range players {
		userName := generateNewUserName()
		scAddress := selectDNSAddressFromUserName(sortedDNSAddresses, userName)
		_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, []byte(scAddress), dnsRegisterValue,
			gasLimit, []byte("register@"+hex.EncodeToString([]byte(userName))))
		userNames[i] = userName
	}

	return userNames
}

func sendRegisterUserNameTxForPlayers(
	players []*integrationTests.TestWalletAccount,
	nodes []*integrationTests.TestProcessorNode,
	sortedDNSAddresses []string,
	dnsRegisterValue *big.Int,
) []string {

	gasLimit := uint64(2000000)
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
	sortedDNSAddresses []string,
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

			bech32c := integrationTests.TestAddressPubkeyConverter
			usernameReportedByNode, err := node.Node.GetUsername(bech32c.Encode(player.Address))
			require.NoError(t, err)
			require.Equal(t, string(hashedUserName), usernameReportedByNode)
		}

		dnsAddress := selectDNSAddressFromUserName(sortedDNSAddresses, userNames[i])
		scQuery := &process.SCQuery{
			CallerAddr: player.Address,
			ScAddress:  []byte(dnsAddress),
			FuncName:   "resolve",
			Arguments:  [][]byte{[]byte(userNames[i])},
		}

		dnsSHId := nodes[0].ShardCoordinator.ComputeId([]byte(dnsAddress))
		for _, node := range nodes {
			if node.ShardCoordinator.SelfId() != dnsSHId {
				continue
			}

			vmOutput, _ := node.SCQueryService.ExecuteQuery(scQuery)

			require.NotNil(t, vmOutput)
			require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
			require.Equal(t, len(vmOutput.ReturnData), 1)
			assert.True(t, bytes.Equal(player.Address, vmOutput.ReturnData[0]))
		}
	}

}

func selectDNSAddressFromUserName(sortedDNSAddresses []string, userName string) string {
	hashedAddr := keccak.Keccak{}.Compute(userName)
	return sortedDNSAddresses[hashedAddr[31]]
}

func generateNewUserName() string {
	return RandStringBytes(10)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
