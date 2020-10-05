package staking

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationtests/frontend/staking")

func TestSignatureOnStaking(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	skHexBuff, pkString, err := core.LoadSkPkFromPemFile("./testdata/key.pem", 0)
	require.Nil(t, err)

	skBuff, err := hex.DecodeString(string(skHexBuff))
	require.Nil(t, err)

	skStaking, err := integrationTests.TestKeyGenForAccounts.PrivateKeyFromByteArray(skBuff)
	require.Nil(t, err)
	pkStaking := skStaking.GeneratePublic()
	pkBuff, err := pkStaking.ToByteArray()
	require.Nil(t, err)

	stakingWalletAccount := &integrationTests.TestWalletAccount{
		SingleSigner:      &ed25519SingleSig.Ed25519Signer{},
		BlockSingleSigner: &singlesig.BlsSingleSigner{},
		SkTxSign:          skStaking,
		PkTxSign:          pkStaking,
		PkTxSignBytes:     pkBuff,
		KeygenTxSign:      integrationTests.TestKeyGenForAccounts,
		PeerSigHandler:    nil,
		Address:           pkBuff,
		Nonce:             0,
		Balance:           big.NewInt(0),
	}

	log.Info("using tx sign pk for staking", "pk", pkString)

	frontendBLSPubkey, err := hex.DecodeString("cbc8c9a6a8d9c874e89eb9366139368ae728bd3eda43f173756537877ba6bca87e01a97b815c9f691df73faa16f66b15603056540aa7252d73fecf05d24cd36b44332a88386788fbdb59d04502e8ecb0132d8ebd3d875be4c83e8b87c55eb901")
	require.Nil(t, err)
	frontendHexSignature := "aabbcc"

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodesWithBLSSigVerifier(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)
	integrationTests.MintAllPlayers(nodes, []*integrationTests.TestWalletAccount{stakingWalletAccount}, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send stake tx and check sender's balance
	var txData string
	genesisBlock := nodes[0].GenesisBlocks[core.MetachainShardId]
	metaBlock := genesisBlock.(*block.MetaBlock)
	nodePrice := big.NewInt(0).Set(metaBlock.EpochStart.Economics.NodePrice)
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())

	pubKey := hex.EncodeToString(frontendBLSPubkey)
	txData = "stake" + "@" + oneEncoded + "@" + pubKey + "@" + frontendHexSignature
	sendTransaction(nodes, stakingWalletAccount, nodePrice, txData)

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	time.Sleep(time.Second)

	testStakingWasDone(t, nodes, frontendBLSPubkey)
}

func sendTransaction(
	nodes []*integrationTests.TestProcessorNode,
	stakingWalletAccount *integrationTests.TestWalletAccount,
	nodePrice *big.Int,
	txData string,
) {
	var senderNode *integrationTests.TestProcessorNode
	for _, n := range nodes {
		if n.ShardCoordinator.ComputeId(stakingWalletAccount.Address) == n.ShardCoordinator.SelfId() {
			senderNode = n
			break
		}
	}

	integrationTests.CreateAndSendTransactionFromWallet(
		senderNode,
		stakingWalletAccount,
		nodePrice,
		vm.AuctionSCAddress,
		txData,
		1,
	)
}

func testStakingWasDone(t *testing.T, nodes []*integrationTests.TestProcessorNode, blsKey []byte) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == core.MetachainShardId {
			checkStakeOnNode(t, n, blsKey)
		}
	}
}

func checkStakeOnNode(t *testing.T, n *integrationTests.TestProcessorNode, blsKey []byte) {
	query := &process.SCQuery{
		ScAddress: vm.StakingSCAddress,
		FuncName:  "getBLSKeyStatus",
		Arguments: [][]byte{blsKey},
	}

	vmOutput, err := n.SCQueryService.ExecuteQuery(query)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.Equal(t, 1, len(vmOutput.ReturnData))
	assert.Equal(t, []byte("staked"), vmOutput.ReturnData[0])
}
