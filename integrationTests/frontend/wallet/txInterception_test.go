package wallet

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

const mintingValue = "100000000"

func TestInterceptedTxWhithoutDataField(t *testing.T) {
	value := big.NewInt(0)
	value.SetString("999", 10)

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		value,
		"erd12dnfhej64s6c56ka369gkyj3hwv5ms0y5rxgsk2k7hkd2vuk7rvqxkalsa",
		"erd1nqxxd22yfgzgulvemred666gdgkzy3qa2qyxwyvvqp6v4ug5h5fs2vlxz4",
		"595dbad31b1ef6e96ecc0ef66f1b51ee7a2308059fd222b7d9c9a709541467454097de68879a48f762afa9d553a997bcb733368c6e4bafc5b3ae535dd5c07f02",
		10,
		100000,
		[]byte("!!!!!"),
		[]byte("test chain ID"),
		integrationTests.MinTransactionVersion,
	)
}

func TestInterceptedTxWhithDataField(t *testing.T) {
	value := big.NewInt(0)
	value.SetString("999", 10)

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		value,
		"erd12dnfhej64s6c56ka369gkyj3hwv5ms0y5rxgsk2k7hkd2vuk7rvqxkalsa",
		"erd14t6l0x27w4d4354sqfm40wuv9p0r49uzl9598eka290x9kws2nvqlkc36j",
		"1e4692dffac8d6c3ca4115676f866e2ded9322b6ba6981853f48d737dd041e8b7255bd9d61636e060a1d9664d97c0f015a26a2cf2a7111a0f1a67c56d54ee902",
		10,
		100000,
		[]byte("data@~`!@#$^&*()_=[]{};'<>?,./|<>><!!!!!"),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
}

// testInterceptedTxFromFrontendGeneratedParams tests that a frontend generated tx will pass through an interceptor
// and ends up in the datapool, concluding the tx is correctly signed and follows our protocol
func testInterceptedTxFromFrontendGeneratedParams(
	t *testing.T,
	frontendNonce uint64,
	frontendValue *big.Int,
	frontendReceiver string,
	frontendSender string,
	frontendSignatureHex string,
	frontendGasPrice uint64,
	frontendGasLimit uint64,
	frontendData []byte,
	chainID []byte,
	version uint32,
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

	node.DataPool.Transactions().RegisterOnAdded(func(key []byte, value interface{}) {
		assert.Equal(t, txHexHash, hex.EncodeToString(key))

		dataRecovered, _ := node.DataPool.Transactions().SearchFirstData(key)
		assert.NotNil(t, dataRecovered)

		txRecovered, ok := dataRecovered.(*transaction.Transaction)
		assert.True(t, ok)

		assert.Equal(t, frontendNonce, txRecovered.Nonce)
		assert.Equal(t, frontendValue, txRecovered.Value)

		sender, _ := integrationTests.TestAddressPubkeyConverter.Decode(frontendSender)
		assert.Equal(t, sender, txRecovered.SndAddr)

		receiver, _ := integrationTests.TestAddressPubkeyConverter.Decode(frontendReceiver)
		assert.Equal(t, receiver, txRecovered.RcvAddr)

		sig, _ := hex.DecodeString(frontendSignatureHex)
		assert.Equal(t, sig, txRecovered.Signature)
		assert.Equal(t, len(frontendData), len(txRecovered.Data))

		chDone <- struct{}{}
	})

	rcvAddrBytes, _ := integrationTests.TestAddressPubkeyConverter.Decode(frontendReceiver)
	sndAddrBytes, _ := integrationTests.TestAddressPubkeyConverter.Decode(frontendSender)
	signatureBytes, _ := hex.DecodeString(frontendSignatureHex)

	integrationTests.MintAddress(node.AccntState, sndAddrBytes, valMinting)

	tx := &transaction.Transaction{
		Nonce:     frontendNonce,
		RcvAddr:   rcvAddrBytes,
		SndAddr:   sndAddrBytes,
		GasPrice:  frontendGasPrice,
		GasLimit:  frontendGasLimit,
		Data:      frontendData,
		Signature: signatureBytes,
		ChainID:   chainID,
		Version:   version,
	}
	tx.Value = big.NewInt(0).Set(frontendValue)
	txHexHash, err = node.SendTransaction(tx)

	assert.Nil(t, err)

	select {
	case <-chDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout getting transaction")
	}
}
