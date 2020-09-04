package attestationContract

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/relayedTx"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
)

func TestRelayedTransactionInMultiShardEnvironmentWithAttestationContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := relayedTx.CreateGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	ownerNode := nodes[0]
	scCode := arwen.GetSCCode("attestation.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	registerValue := big.NewInt(100)
	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		200000,
		make([]byte, 32),
		[]byte(arwen.CreateDeployTxData(scCode)+"@"+hex.EncodeToString(registerValue.Bytes())+"@"+hex.EncodeToString(relayer.Address)+"@"+"ababab"),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
	time.Sleep(time.Second)

	registerVMGas := uint64(100000)
	savePublicInfoVMGas := uint64(100000)
	attestVMGas := uint64(100000)

	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	uniqueIDs := make([]string, len(players))
	for i, player := range players {
		uniqueIDs[i] = core.UniqueIdentifier()
		_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, registerValue,
			registerVMGas, []byte("register@"+hex.EncodeToString([]byte(uniqueIDs[i]))))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := int64(10)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	cryptoHook := hooks.NewVMCryptoHook()
	privateInfos := make([]string, len(players))
	for i := range players {
		privateInfos[i] = core.UniqueIdentifier()
		publicInfo, _ := cryptoHook.Keccak256([]byte(privateInfos[i]))
		relayedTx.CreateAndSendSimpleTransaction(nodes, relayer, scAddress, big.NewInt(0), savePublicInfoVMGas,
			[]byte("savePublicInfo@"+hex.EncodeToString([]byte(uniqueIDs[i]))+"@"+hex.EncodeToString(publicInfo)))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagate := int64(5)
	for i := int64(0); i <= nrRoundsToPropagate; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	for i, player := range players {
		_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0), attestVMGas,
			[]byte("attest@"+hex.EncodeToString([]byte(uniqueIDs[i]))+"@"+hex.EncodeToString([]byte(privateInfos[i]))))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard = int64(20)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	for i, player := range players {
		relayedTx.CheckAttestedPublicKeys(t, ownerNode, scAddress, []byte(uniqueIDs[i]), player.Address)
	}
}
