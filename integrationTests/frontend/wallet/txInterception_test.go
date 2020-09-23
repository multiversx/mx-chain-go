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

func TestInterceptedTxWithoutDataField(t *testing.T) {
	value := big.NewInt(0)
	value.SetString("999", 10)

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		value,
		"erd12dnfhej64s6c56ka369gkyj3hwv5ms0y5rxgsk2k7hkd2vuk7rvqxkalsa",
		"erd15xa6lw0g8uvnkl82p9p9ks5guvp02twuwky4m936lhtup4cppejsg2rv0g",
		"9f0e3bccbe1815580e2f8db150405e2f88255aaf801cde204af8728ccd4d43d1ab82188a24050e00d1cc21f85eab251014e16984774957ce06d2060fd77e9f0f",
		10,
		100000,
		[]byte(""),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
}

func TestInterceptedTxWithDataField(t *testing.T) {
	value := big.NewInt(0)
	value.SetString("999", 10)

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		value,
		"erd12dnfhej64s6c56ka369gkyj3hwv5ms0y5rxgsk2k7hkd2vuk7rvqxkalsa",
		"erd1zc858kk5ln3pnkwfw4lwpx7ssj890vkdtg47tzw4uvkcmja3sv7sf8dsg3",
		"7486cceda06922b047e0601576dec8c5bea8b29b89fe7ee268a27550e0dc9acde9dd3214bb4b64f6d273a95b2c5be8c9f83a57d8c35e74c11e4f376e60848801",
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
