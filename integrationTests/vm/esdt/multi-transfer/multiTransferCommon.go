package multitransfer

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	testVm "github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/require"
)

const NR_ROUNDS_CROSS_SHARD = 15
const NR_ROUNDS_SAME_SHARD = 1

type esdtTransfer struct {
	tokenIdentifier string
	nonce           int64
	amount          int64
}

func issueFungibleToken(t *testing.T,
	net *integrationTests.TestNetwork,
	issuerNode *integrationTests.TestProcessorNode,
	ticker string,
	initialSupply int64) string {

	issuerAddress := issuerNode.OwnAccount.Address

	tokenName := "token"
	issuePrice := big.NewInt(1000)
	txData := txDataBuilder.NewBuilder()
	txData.IssueESDT(tokenName, ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true)

	integrationTests.CreateAndSendTransaction(
		issuerNode,
		net.Nodes,
		issuePrice,
		vm.ESDTSCAddress,
		txData.ToString(), core.MinMetaTxExtraGasCost)
	waitForOperationCompletion(net, NR_ROUNDS_CROSS_SHARD)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(net.Nodes, []byte(ticker)))

	esdt.CheckAddressHasTokens(t, issuerAddress, net.Nodes,
		tokenIdentifier, 0, initialSupply)

	return tokenIdentifier
}

func issueNft(
	net *integrationTests.TestNetwork,
	issuerNode *integrationTests.TestProcessorNode,
	ticker string,
	semiFungible bool) string {

	tokenName := "token"
	issuePrice := big.NewInt(1000)

	txData := txDataBuilder.NewBuilder()

	issueFunc := "issueNonFungible"
	if semiFungible {
		issueFunc = "issueSemiFungible"
	}

	txData.Func(issueFunc).Str(tokenName).Str(ticker)
	txData.CanFreeze(false).CanWipe(false).CanPause(false).CanTransferNFTCreateRole(true)

	integrationTests.CreateAndSendTransaction(
		issuerNode,
		net.Nodes,
		issuePrice,
		vm.ESDTSCAddress,
		txData.ToString(),
		core.MinMetaTxExtraGasCost)
	waitForOperationCompletion(net, NR_ROUNDS_CROSS_SHARD)

	issuerAddress := issuerNode.OwnAccount.Address
	tokenIdentifier := string(integrationTests.GetTokenIdentifier(net.Nodes, []byte(ticker)))

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}
	if semiFungible {
		roles = append(roles, []byte(core.ESDTRoleNFTAddQuantity))
	}

	setLocalRoles(net, issuerNode, issuerAddress, tokenIdentifier, roles)

	return tokenIdentifier
}

func setLocalRoles(
	net *integrationTests.TestNetwork,
	issuerNode *integrationTests.TestProcessorNode,
	addrForRole []byte,
	tokenIdentifier string,
	roles [][]byte,
) {
	txData := "setSpecialRole" +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(addrForRole)

	for _, role := range roles {
		txData += "@" + hex.EncodeToString(role)
	}

	integrationTests.CreateAndSendTransaction(
		issuerNode,
		net.Nodes,
		big.NewInt(0),
		vm.ESDTSCAddress,
		txData,
		core.MinMetaTxExtraGasCost)
	waitForOperationCompletion(net, NR_ROUNDS_CROSS_SHARD)
}

func createSFT(
	t *testing.T,
	net *integrationTests.TestNetwork,
	issuerNode *integrationTests.TestProcessorNode,
	tokenIdentifier string,
	createdTokenNonce int64,
	initialSupply int64) {

	issuerAddress := issuerNode.OwnAccount.Address

	tokenName := "token"
	royalties := big.NewInt(0)
	hash := "someHash"
	attributes := "cool nft"
	uri := "www.my-cool-nfts.com"

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionESDTNFTCreate)
	txData.Str(tokenIdentifier)
	txData.Int64(initialSupply)
	txData.Str(tokenName)
	txData.BigInt(royalties)
	txData.Str(hash)
	txData.Str(attributes)
	txData.Str(uri)

	integrationTests.CreateAndSendTransaction(
		issuerNode,
		net.Nodes,
		big.NewInt(0),
		issuerAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit)
	waitForOperationCompletion(net, NR_ROUNDS_SAME_SHARD)

	esdt.CheckAddressHasTokens(t, issuerAddress, net.Nodes,
		tokenIdentifier, createdTokenNonce, initialSupply)
}

func createNFT(
	t *testing.T,
	net *integrationTests.TestNetwork,
	issuerNode *integrationTests.TestProcessorNode,
	tokenIdentifier string,
	createdTokenNonce int64) {

	createSFT(t, net, issuerNode, tokenIdentifier, createdTokenNonce, 1)
}

func buildEsdtMultiTransferTxData(
	receiverAddress []byte,
	transfers []*esdtTransfer,
	endpointName string,
	arguments ...[]byte) string {

	nrTransfers := len(transfers)

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(receiverAddress)
	txData.Int(nrTransfers)

	for _, transfer := range transfers {
		txData.Str(transfer.tokenIdentifier)
		txData.Int64(transfer.nonce)
		txData.Int64(transfer.amount)
	}

	if len(endpointName) > 0 {
		txData.Str(endpointName)

		for _, arg := range arguments {
			txData.Bytes(arg)
		}
	}

	return txData.ToString()
}

func waitForOperationCompletion(net *integrationTests.TestNetwork, roundsToWait int) {
	time.Sleep(time.Second)
	net.Steps(roundsToWait)
}

func multiTransferToVault(
	t *testing.T,
	net *integrationTests.TestNetwork,
	senderNode *integrationTests.TestProcessorNode,
	vaultScAddress []byte,
	transfers []*esdtTransfer,
	nrRoundsToWait int,
	userBalances map[string]map[int64]int64,
	scBalances map[string]map[int64]int64) {

	acceptMultiTransferEndpointName := "accept_funds_multi_transfer"
	senderAddress := senderNode.OwnAccount.Address

	txData := buildEsdtMultiTransferTxData(vaultScAddress,
		transfers,
		acceptMultiTransferEndpointName,
	)

	integrationTests.CreateAndSendTransaction(
		senderNode,
		net.Nodes,
		big.NewInt(0),
		senderAddress,
		txData,
		integrationTests.AdditionalGasLimit,
	)
	waitForOperationCompletion(net, nrRoundsToWait)

	// update expected balances after transfers
	for _, transfer := range transfers {
		userBalances[transfer.tokenIdentifier][transfer.nonce] -= transfer.amount
		scBalances[transfer.tokenIdentifier][transfer.nonce] += transfer.amount
	}

	// check expected vs actual values
	for _, transfer := range transfers {
		expectedUserBalance := userBalances[transfer.tokenIdentifier][transfer.nonce]
		expectedScBalance := scBalances[transfer.tokenIdentifier][transfer.nonce]

		esdt.CheckAddressHasTokens(t, senderAddress, net.Nodes,
			transfer.tokenIdentifier, transfer.nonce, expectedUserBalance)
		esdt.CheckAddressHasTokens(t, vaultScAddress, net.Nodes,
			transfer.tokenIdentifier, transfer.nonce, expectedScBalance)
	}
}

func deployNonPayableSmartContract(
	t *testing.T,
	net *integrationTests.TestNetwork,
	fileName string,
) []byte {
	scCode := arwen.GetSCCode(fileName)

	deployerNode := net.NodesSharded[0][0]
	scAddress, _ := deployerNode.BlockchainHook.NewAddress(
		deployerNode.OwnAccount.Address,
		deployerNode.OwnAccount.Nonce,
		vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		deployerNode,
		net.Nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxDataNonPayable(scCode),
		integrationTests.AdditionalGasLimit,
	)
	waitForOperationCompletion(net, 4)

	_, err := deployerNode.AccntState.GetExistingAccount(scAddress)
	require.Nil(t, err)

	return scAddress
}
