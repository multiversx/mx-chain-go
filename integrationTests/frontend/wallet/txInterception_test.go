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

func TestInterceptedTxFromFrontendLargeValue(t *testing.T) {
	value := big.NewInt(0)
	value.SetString("1000999999999999999999991234", 10)

	fmt.Println(value.Text(10))
	fmt.Println(value.Text(16))

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		value,
		"c2981474860ebd42f9da812a41dcace8a0c2fdac52e3a66a45603821ca4c6d43",
		"c2981474860ebd42f9da812a41dcace8a0c2fdac52e3a66a45603821ca4c6d43",
		"469d44b058faadb56cabbc696f2a0f5c9d4a361b3432c37135d6216feb03fcce890ebc3b98d1506be0cf88f5f22ad533a90386b2211aaad6df32a41be4b01e09",
		10,
		1002,
		"de",
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

	chDone := make(chan struct{})

	maxShards := uint32(1)
	nodeShardId := uint32(0)
	txSignPrivKeyShardId := uint32(0)
	initialNodeAddr := "nodeAddr"
	valMinting := big.NewInt(0).Set(frontendValue)
	valMinting.Mul(valMinting, big.NewInt(2))

	node := integrationTests.NewTestProcessorNode(
		maxShards,
		nodeShardId,
		txSignPrivKeyShardId,
		initialNodeAddr,
	)

	txHexHash := ""

	err := node.SetAccountNonce(uint64(0))
	assert.Nil(t, err)

	node.ShardDataPool.Transactions().RegisterHandler(func(key []byte) {
		assert.Equal(t, txHexHash, hex.EncodeToString(key))

		dataRecovered, _ := node.ShardDataPool.Transactions().SearchFirstData(key)
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
	tx.SetValue(frontendValue)
	txHexHash, err = node.SendTransaction(tx)

	assert.Nil(t, err)

	select {
	case <-chDone:
	case <-time.After(time.Second * 2):
		assert.Fail(t, "timeout getting transaction")
	}
}
