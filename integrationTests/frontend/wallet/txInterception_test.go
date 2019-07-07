package wallet

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestInterceptedTxFromFrontendGeneratedParamsWithoutData(t *testing.T) {
	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		big.NewInt(10),
		"53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8",
		"fe73b8960894941bcf100f7378dba2a6fa2591343413710073c2515817b27dc5",
		"f2ae2ad6585f3b44bbbe84f93c3c5ec04a53799d24c04a1dd519666f2cd3dc3d7fbe6c75550b0eb3567fdc0708a8534ae3e5393d0dd9e03c70972f2e716a7007",
		0,
		0,
		"",
	)
}

func TestInterceptedTxFromFrontendGeneratedParams(t *testing.T) {
	t.Skip("repair this test")

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		big.NewInt(10),
		"53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8",
		"3d4356c1ed18a3f77650be955019447e5a851f7cd855ff727bd2d54b63012a9d",
		"80c7943ac75727fc2250cbbd1734a36474ddddd3f121da9f9e98f0ca8ab8789c32ac07435bafcf64e8173e06e3863021af2a4be59d364dc6b8b3106adc14400f",
		0,
		0,
		"53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8",
	)
}

func TestInterceptedTxFromFrontendGeneratedParamsAllParams(t *testing.T) {
	t.Skip("repair this test")

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		big.NewInt(10),
		"53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8",
		"6afb8018dcc5a53d22d4dcdda39ceaf25dafd1ea353a9bbe12073057f4e6d262",
		"822b290b5100d3e403b2a4e39ff8192c2298a17d45b7f903da410ddbdeef2f08ebb5f4029757d1cc5a2820de0532f786d30620a2dfc64adcfe637bb8cb7bb207",
		10,
		1000,
		"53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8",
	)
}

func TestInterceptedTxFromFrontendGeneratedParamsAllParams2(t *testing.T) {
	testInterceptedTxFromFrontendGeneratedParams(
		t,
		12,
		big.NewInt(2),
		"943643524936191d1c5627e044f7b5e4ca559c7d0ba1c2b85d1b2e6c299ebcd8",
		"943643524936191d1c5627e044f7b5e4ca559c7d0ba1c2b85d1b2e6c299ebcd8",
		"1ef83bae21227e93e9717f45a4ec34e3f5c6a110e31dfa438ac2b8c1f5459e5167fd8424d1dfa6de59756437fe599def6872217ddad5717fe61a41853606450c",
		1,
		10000,
		"aa@dd@cc",
	)
}

func TestInterceptedTxFromFrontendGeneratedParamsGasPriceGasLimitNoData(t *testing.T) {
	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		big.NewInt(10),
		"53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8",
		"6afb8018dcc5a53d22d4dcdda39ceaf25dafd1ea353a9bbe12073057f4e6d262",
		"1d96166ecd6cae86797046126b64028099fcd026a37a82c4bdd19700bd49828069a822fb5453e0b32f66ed895d4f162af35ea8aca862af498e2831c596250e03",
		10,
		1000,
		"",
	)
}

// testInterceptedTxFromFrontendGeneratedParams tests that a frontend generated tx will pass through an interceptor
// and ends up in the datapool, concluding the tx is correctly signed and follows our protocol
func testInterceptedTxFromFrontendGeneratedParams(
	t *testing.T,
	frontendNonce uint64,
	frontendValue *big.Int,
	frontendReceiverHex string,
	frontendSenderHex string,
	frontendSignature string,
	frontendGasPrice uint64,
	frontendGasLimit uint64,
	frontendData string,
) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	_ = nodeAccount.(*state.Account).SetNonceWithJournal(startingNonce)
	_, _ = accntAdapter.Commit()

	chDone := make(chan struct{})

	var err error
	txHexHash := ""

	dPool.Transactions().RegisterHandler(func(key []byte) {
		assert.Equal(t, txHexHash, hex.EncodeToString(key))

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
		assert.Equal(t, txRecovered.Data, frontendData)

		chDone <- struct{}{}
	})

	sig, _ := hex.DecodeString(frontendSignature)
	txHexHash, err = n.SendTransaction(frontendNonce, frontendSenderHex, frontendReceiverHex,
		frontendValue, frontendGasPrice, frontendGasLimit, frontendData, sig)
	assert.Nil(t, err)

	select {
	case <-chDone:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "timeout getting transaction")
	}
}
