package relayedTx

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := createGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_ = createAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			_ = createAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, integrationTests.MinTxGasLimit, []byte(""))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	receiver1 := getUserAccount(nodes, receiverAddress1)
	receiver2 := getUserAccount(nodes, receiverAddress2)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	assert.Equal(t, receiver1.GetBalance().Cmp(finalBalance), 0)
	assert.Equal(t, receiver2.GetBalance().Cmp(finalBalance), 0)

	players = append(players, relayer)
	checkPlayerBalances(t, nodes, players)
}

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTxButWrongNonce(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := createGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	totalFees := big.NewInt(0)
	relayerInitialValue := big.NewInt(0).Set(relayer.Balance)
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			player.Nonce += 1
			relayerTx := createAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			totalFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(relayerTx.GetGasPrice()), big.NewInt(0).SetUint64(relayerTx.GetGasLimit()))
			totalFees.Add(totalFees, totalFee)
			relayerTx = createAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			totalFee = big.NewInt(0).Mul(big.NewInt(0).SetUint64(relayerTx.GetGasPrice()), big.NewInt(0).SetUint64(relayerTx.GetGasLimit()))
			totalFees.Add(totalFees, totalFee)
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	receiver1 := getUserAccount(nodes, receiverAddress1)
	receiver2 := getUserAccount(nodes, receiverAddress2)

	assert.True(t, check.IfNil(receiver1))
	assert.True(t, check.IfNil(receiver2))

	for _, player := range players {
		account := getUserAccount(nodes, player.Address)
		assert.True(t, account.GetBalance().Cmp(big.NewInt(0)) == 0)
		assert.Equal(t, uint64(nrRoundsToTest)*2, account.GetNonce())
	}

	expectedBalance := big.NewInt(0).Sub(relayerInitialValue, totalFees)
	relayerAccount := getUserAccount(nodes, relayer.Address)
	assert.True(t, relayerAccount.GetBalance().Cmp(expectedBalance) == 0)
}

func TestRelayedTransactionInMultiShardEnvironmentWithSmartContractTX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := createGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	ownerNode := nodes[0]
	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	scCode := arwen.GetSCCode("../../vm/arwen/testdata/erc20-c-03/wrc20_arwen.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		integrationTests.MaxGasLimitPerBlock-1,
		make([]byte, 32),
		[]byte(arwen.CreateDeployTxData(scCode)+"@"+initialSupply),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	transferTokenVMGas := uint64(6960)
	transferTokenBaseGas := ownerNode.EconomicsData.ComputeGasLimit(&transaction.Transaction{Data: []byte("transferToken@" + hex.EncodeToString(receiverAddress1) + "@00" + hex.EncodeToString(sendValue.Bytes()))})
	transferTokenFullGas := transferTokenBaseGas + transferTokenVMGas

	initialTokenSupply := big.NewInt(1000000000)
	initialPlusForGas := uint64(1000)
	for _, player := range players {
		integrationTests.CreateAndSendTransactionWithGasLimit(
			ownerNode,
			big.NewInt(0),
			transferTokenFullGas+initialPlusForGas,
			scAddress,
			[]byte("transferToken@"+hex.EncodeToString(player.Address)+"@00"+hex.EncodeToString(initialTokenSupply.Bytes())),
			integrationTests.ChainID,
			integrationTests.MinTransactionVersion,
		)
	}
	time.Sleep(time.Second)

	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		for _, player := range players {
			_ = createAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress1)+"@00"+hex.EncodeToString(sendValue.Bytes())))
			_ = createAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress2)+"@00"+hex.EncodeToString(sendValue.Bytes())))
		}

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)

	checkSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)
	checkSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)

	players = append(players, relayer)
	checkPlayerBalances(t, nodes, players)
}

func TestRelayedTransactionInMultiShardEnvironmentWithESDTTX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := createGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	///////////------- send token issue
	tokenName := "robertWhyNot"
	issuePrice := big.NewInt(1000)
	initalSupply := big.NewInt(10000000000)
	tokenIssuer := nodes[0]
	txData := "issue" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(initalSupply.Bytes())
	integrationTests.CreateAndSendTransaction(tokenIssuer, issuePrice, vm.ESDTSCAddress, txData, integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := int64(10)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, initalSupply)

	/////////------ send tx to players
	valueToTopUp := big.NewInt(100000000)
	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(valueToTopUp.Bytes())
	for _, player := range players {
		integrationTests.CreateAndSendTransaction(tokenIssuer, big.NewInt(0), player.Address, txData, integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(sendValue.Bytes())
	transferTokenESDTGas := uint64(1)
	transferTokenBaseGas := tokenIssuer.EconomicsData.ComputeGasLimit(&transaction.Transaction{Data: []byte(txData)})
	transferTokenFullGas := transferTokenBaseGas + transferTokenESDTGas
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_ = createAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, big.NewInt(0), transferTokenFullGas, []byte(txData))
			_ = createAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, big.NewInt(0), transferTokenFullGas, []byte(txData))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	nrRoundsToPropagateMultiShard = int64(20)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	checkAddressHasESDTTokens(t, receiverAddress1, nodes, tokenName, finalBalance)
	checkAddressHasESDTTokens(t, receiverAddress2, nodes, tokenName, finalBalance)

	players = append(players, relayer)
	checkPlayerBalances(t, nodes, players)
}

func TestRelayedTransactionInMultiShardEnvironmentWithAttestationContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := createGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, node := range nodes {
		node.EconomicsData.SetMaxGasLimitPerBlock(1500000000)
	}

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
		integrationTests.MaxGasLimitPerBlock-1,
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
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

	uniqueIDs := make([]string, len(players))
	for i, player := range players {
		uniqueIDs[i] = core.UniqueIdentifier()
		_ = createAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, registerValue,
			registerVMGas, []byte("register@"+hex.EncodeToString([]byte(uniqueIDs[i]))))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := int64(10)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	cryptoHook := hooks.NewVMCryptoHook()
	privateInfos := make([]string, len(players))
	for i := range players {
		privateInfos[i] = core.UniqueIdentifier()
		publicInfo, _ := cryptoHook.Keccak256([]byte(privateInfos[i]))
		createAndSendSimpleTransaction(nodes, relayer, scAddress, big.NewInt(0), savePublicInfoVMGas,
			[]byte("savePublicInfo@"+hex.EncodeToString([]byte(uniqueIDs[i]))+"@"+hex.EncodeToString(publicInfo)))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagate := int64(5)
	for i := int64(0); i <= nrRoundsToPropagate; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	for i, player := range players {
		_ = createAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0), attestVMGas,
			[]byte("attest@"+hex.EncodeToString([]byte(uniqueIDs[i]))+"@"+hex.EncodeToString([]byte(privateInfos[i]))))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard = int64(20)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	for i, player := range players {
		checkAttestedPublicKeys(t, ownerNode, scAddress, []byte(uniqueIDs[i]), player.Address)
	}
}

func checkAttestedPublicKeys(
	t *testing.T,
	node *integrationTests.TestProcessorNode,
	scAddress []byte,
	obfuscatedData []byte,
	userAddress []byte,
) {
	scQuery := node.SCQueryService
	vmOutput, err := scQuery.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "getPublicKey",
		Arguments: [][]byte{obfuscatedData},
	})
	require.Nil(t, err)
	require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
	require.Equal(t, vmOutput.ReturnData[0], userAddress)
}

func checkSCBalance(t *testing.T, node *integrationTests.TestProcessorNode, scAddress []byte, userAddress []byte, balance *big.Int) {
	scQuery := node.SCQueryService
	vmOutput, err := scQuery.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "balanceOf",
		Arguments: [][]byte{userAddress},
	})
	assert.Nil(t, err)
	actualBalance := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
	assert.Equal(t, actualBalance.Cmp(balance), 0)
}

func checkPlayerBalances(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount) {
	for _, player := range players {
		userAcc := getUserAccount(nodes, player.Address)
		assert.Equal(t, userAcc.GetBalance().Cmp(player.Balance), 0)
		assert.Equal(t, userAcc.GetNonce(), player.Nonce)
	}
}

func getUserAccount(
	nodes []*integrationTests.TestProcessorNode,
	address []byte,
) state.UserAccountHandler {
	shardID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == shardID {
			acc, _ := node.AccntState.GetExistingAccount(address)
			if check.IfNil(acc) {
				return nil
			}
			userAcc := acc.(state.UserAccountHandler)
			return userAcc
		}
	}
	return nil
}

func createUserTx(
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    player.Nonce,
		Value:    big.NewInt(0).Set(value),
		RcvAddr:  rcvAddr,
		SndAddr:  player.Address,
		GasPrice: integrationTests.MinTxGasPrice,
		GasLimit: gasLimit,
		Data:     txData,
		ChainID:  integrationTests.ChainID,
		Version:  integrationTests.MinTransactionVersion,
	}
	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	tx.Signature, _ = player.SingleSigner.Sign(player.SkTxSign, txBuff)
	player.Nonce++
	return tx
}

func createRelayedTx(
	feeHandler process.FeeHandler,
	relayer *integrationTests.TestWalletAccount,
	userTx *transaction.Transaction,
) *transaction.Transaction {

	userTxMarshaled, _ := integrationTests.TestTxSignMarshalizer.Marshal(userTx)
	txData := core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshaled)
	tx := &transaction.Transaction{
		Nonce:    relayer.Nonce,
		Value:    big.NewInt(0).Set(userTx.Value),
		RcvAddr:  userTx.SndAddr,
		SndAddr:  relayer.Address,
		GasPrice: integrationTests.MinTxGasPrice,
		Data:     []byte(txData),
		ChainID:  userTx.ChainID,
		Version:  userTx.Version,
	}
	gasLimit := feeHandler.ComputeGasLimit(tx)
	tx.GasLimit = userTx.GasLimit + gasLimit

	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	tx.Signature, _ = relayer.SingleSigner.Sign(relayer.SkTxSign, txBuff)
	relayer.Nonce++
	txFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(tx.GasLimit), big.NewInt(0).SetUint64(tx.GasPrice))
	relayer.Balance.Sub(relayer.Balance, txFee)
	relayer.Balance.Sub(relayer.Balance, tx.Value)

	return tx
}

func createAndSendRelayedAndUserTx(
	nodes []*integrationTests.TestProcessorNode,
	relayer *integrationTests.TestWalletAccount,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) *transaction.Transaction {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address)

	userTx := createUserTx(player, rcvAddr, value, gasLimit, txData)
	relayedTx := createRelayedTx(txDispatcherNode.EconomicsData, relayer, userTx)

	_, err := txDispatcherNode.SendTransaction(relayedTx)
	if err != nil {
		fmt.Println(err.Error())
	}

	return relayedTx
}

func createAndSendSimpleTransaction(
	nodes []*integrationTests.TestProcessorNode,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address)

	userTx := createUserTx(player, rcvAddr, value, gasLimit, txData)
	_, err := txDispatcherNode.SendTransaction(userTx)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func getNodeWithinSameShardAsPlayer(
	nodes []*integrationTests.TestProcessorNode,
	player []byte,
) *integrationTests.TestProcessorNode {
	nodeWithCaller := nodes[0]
	playerShId := nodeWithCaller.ShardCoordinator.ComputeId(player)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == playerShId {
			nodeWithCaller = node
			break
		}
	}

	return nodeWithCaller
}

func checkAddressHasESDTTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value *big.Int,
) {
	userAcc := getUserAccount(nodes, address)

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenName)
	esdtData, err := getESDTDataFromKey(userAcc, tokenKey)
	assert.Nil(t, err)

	assert.Equal(t, esdtData.Value.Cmp(value), 0)
}

func getESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*builtInFunctions.ESDigitalToken, error) {
	esdtData := &builtInFunctions.ESDigitalToken{Value: big.NewInt(0)}
	marshaledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil {
		return esdtData, nil
	}

	err = integrationTests.TestMarshalizer.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}

func createGeneralSetupForRelayTxTest() ([]*integrationTests.TestProcessorNode, []int, []*integrationTests.TestWalletAccount, *integrationTests.TestWalletAccount, p2p.Messenger) {
	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
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

	initialVal := big.NewInt(1000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	numPlayers := 5
	numShards := nodes[0].ShardCoordinator.NumberOfShards()
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		shardId := uint32(i) % numShards
		players[i] = integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, shardId)
	}

	relayerAccount := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, 0)
	integrationTests.MintAllPlayers(nodes, []*integrationTests.TestWalletAccount{relayerAccount}, initialVal)

	return nodes, idxProposers, players, relayerAccount, advertiser
}
