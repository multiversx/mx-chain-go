package wallet

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

// TestInterceptedTxFromFrontendGeneratedParams tests that a frontend generated tx will pass through an interceptor
// and ends up in the datapool, concluding the tx is correctly signed and follows our protocol
func TestInterceptedTxFromFrontendGeneratedParams(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	frontendNonce := uint64(0)
	frontendValue := big.NewInt(10)
	frontendReceiverHex := "53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8"
	frontendSenderHex := "fe73b8960894941bcf100f7378dba2a6fa2591343413710073c2515817b27dc5"
	frontendSignature := "f2ae2ad6585f3b44bbbe84f93c3c5ec04a53799d24c04a1dd519666f2cd3dc3d7fbe6c75550b0eb3567fdc0708a8534ae3e5393d0dd9e03c70972f2e716a7007"

	dPool := createTestDataPool()
	startingNonce := uint64(0)

	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	accntAdapter := createAccountsDB()

	shardCoordinator := &sharding.OneShardCoordinator{}

	n, _, sk, _ := createNetNode(dPool, accntAdapter, shardCoordinator)

	//set the account's nonce to startingNonce
	nodePubKeyBytes, _ := sk.GeneratePublic().ToByteArray()
	nodeAddress, _ := addrConverter.CreateAddressFromPublicKeyBytes(nodePubKeyBytes)
	nodeAccount, _ := accntAdapter.GetAccountWithJournal(nodeAddress)
	nodeAccount.(*state.Account).SetNonceWithJournal(startingNonce)
	accntAdapter.Commit()

	chDone := make(chan struct{})

	dPool.Transactions().RegisterHandler(func(key []byte) {
		dataRecovered, _ := dPool.Transactions().SearchFirstData(key)
		assert.NotNil(t, dataRecovered)

		txRecovered, ok := dataRecovered.(*transaction.Transaction)
		assert.True(t, ok)

		assert.Equal(t, txRecovered.Nonce, frontendNonce)
		assert.Equal(t, txRecovered.Value, frontendValue)

		sender, _ := hex.DecodeString(frontendSenderHex)
		assert.Equal(t, txRecovered.SndAddr, sender)

		receiver, _ := hex.DecodeString(frontendReceiverHex)
		assert.Equal(t, txRecovered.RcvAddr, receiver)

		sig, _ := hex.DecodeString(frontendSignature)
		assert.Equal(t, txRecovered.Signature, sig)

		chDone <- struct{}{}
	})

	sig, _ := hex.DecodeString(frontendSignature)
	n.SendTransaction(frontendNonce, frontendSenderHex, frontendReceiverHex, frontendValue, "", sig)

	select {
	case <-chDone:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "timeout getting transaction")
	}
}
