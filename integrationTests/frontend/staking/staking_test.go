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

	frontendBLSPubkey, err := hex.DecodeString("309befb6387288380edda61ce174b12d42ad161d19361dfcf7e61e6a4e812caf07e45a5a1c5c1e6e1f2f4d84d794dc16d9c9db0088397d85002194b773c30a8b7839324b3b80d9b8510fe53385ba7b7383c96a4c07810db31d84b0feefafbd03")
	require.Nil(t, err)
	frontendHexSignature := "17b1f945404c0c98d2e69a576f3635f4ebe77cd396561566afb969333b0da053e7485b61ef10311f512e3ec2f351ee95"

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodesWithBLSSigVerifier(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
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
	integrationTests.PlayerSendsTransaction(
		nodes,
		stakingWalletAccount,
		vm.ValidatorSCAddress,
		nodePrice,
		txData,
		integrationTests.MinTxGasLimit+uint64(len(txData))+1+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	time.Sleep(time.Second)

	testStakingWasDone(t, nodes, frontendBLSPubkey)
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
