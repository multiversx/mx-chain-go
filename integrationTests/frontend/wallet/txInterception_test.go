package wallet

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

const mintingValue = "100000000"

func TestInterceptedTxFromFrontendLargeValue(t *testing.T) {
	value := big.NewInt(0)
	value.SetString("777", 10)

	fmt.Println(value.Text(10))
	fmt.Println(value.Text(16))

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		value,
		"53669be65aac358a6add8e8a8b1251bb994dc1e4a0cc885956f5ecd53396f0d8",
		"2d7aa683fbb37eafc2426bfe63e1c20aa5872ee4627c51b6789f41bfb8d31fdb",
		"a18a6c6647d10a579acd7e39258f38cee4cd36998ae12edf4e884066231b00e18d792cc14ece72d3ac6fb26281c5419b1ec9736291d1c9fbb312ee2a730c8103",
		10,
		100000,
		[]byte("a@b@c!!$%^<>#!"),
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
	frontendData []byte,
) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chDone := make(chan struct{})

	maxShards := uint32(1)
	nodeShardId := uint32(0)
	txSignPrivKeyShardId := uint32(0)
	initialNodeAddr := "nodeAddr"
	valMinting, _ := big.NewInt(0).SetString(mintingValue, 10)
	valMinting.Mul(valMinting, big.NewInt(5))

	node := integrationTests.NewTestProcessorNode(
		maxShards,
		nodeShardId,
		txSignPrivKeyShardId,
		initialNodeAddr,
	)

	txHexHash := ""

	err := node.SetAccountNonce(uint64(0))
	assert.Nil(t, err)

	node.DataPool.Transactions().RegisterHandler(func(key []byte) {
		assert.Equal(t, txHexHash, hex.EncodeToString(key))

		dataRecovered, _ := node.DataPool.Transactions().SearchFirstData(key)
		assert.NotNil(t, dataRecovered)

		txRecovered, ok := dataRecovered.(*transaction.Transaction)
		assert.True(t, ok)

		assert.Equal(t, frontendNonce, txRecovered.Nonce)
		assert.Equal(t, frontendValue, txRecovered.Value)

		sender, _ := hex.DecodeString(frontendSenderHex)
		assert.Equal(t, sender, txRecovered.SndAddr)

		receiver, _ := hex.DecodeString(frontendReceiverHex)
		assert.Equal(t, receiver, txRecovered.RcvAddr)

		sig, _ := hex.DecodeString(frontendSignature)
		assert.Equal(t, sig, txRecovered.Signature)
		assert.Equal(t, frontendData, txRecovered.Data)

		chDone <- struct{}{}
	})

	rcvAddrBytes, _ := hex.DecodeString(frontendReceiverHex)
	sndAddrBytes, _ := hex.DecodeString(frontendSenderHex)
	signatureBytes, _ := hex.DecodeString(frontendSignature)

	integrationTests.MintAddress(node.AccntState, sndAddrBytes, valMinting)

	tx := &transaction.Transaction{
		Nonce:     frontendNonce,
		RcvAddr:   rcvAddrBytes,
		SndAddr:   sndAddrBytes,
		GasPrice:  frontendGasPrice,
		GasLimit:  frontendGasLimit,
		Data:      frontendData,
		Signature: signatureBytes,
	}
	tx.SetValue(new(big.Int).Set(frontendValue))
	txHexHash, err = node.SendTransaction(tx)

	assert.Nil(t, err)

	select {
	case <-chDone:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "timeout getting transaction")
	}
}
