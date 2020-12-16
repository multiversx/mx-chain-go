package wallet

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		"erd1l20m7kzfht5rhdnd4zvqr82egk7m4nvv3zk06yw82zqmrt9kf0zsf9esqq",
		"394c6f1375f6511dd281465fb9dd7caf013b6512a8f8ac278bbe2151cbded89da28bd539bc1c1c7884835742712c826900c092edb24ac02de9015f0f494f6c0a",
		10,
		100000,
		[]byte(""),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
		0,
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
		"erd1vjdll4xqtyll7nx3xx83audv27k5ktppvcc8kddn2yhxhwrjadgqlvtz8w",
		"258edbbec6f7b67f6747340da09a5f849fe9fde29758097e06684d261e5d92d3abafbe9b2e2879226d738b2223a82468642e9342032ff69798de69b0fd6ca304",
		10,
		100000,
		[]byte("data@~`!@#$^&*()_=[]{};'<>?,./|<>><!!!!!"),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
		0,
	)
}

func TestInterceptedTxWithSigningOverTxHash(t *testing.T) {
	value := big.NewInt(0)
	value.SetString("1000000000000000000", 10)

	testInterceptedTxFromFrontendGeneratedParams(
		t,
		1,
		value,
		"erd1ez0puv8mqsulwllnavfygfzqe5zveeqjwpr6dsm8egchkf449kjqf8udu6",
		"erd1ez0puv8mqsulwllnavfygfzqe5zveeqjwpr6dsm8egchkf449kjqf8udu6",
		"89cb10cafb75040d704b66610990c2ec7f6393e8da7ac867b1db417a9c9a3340947d5139e23ef14ebe80fd33ce352458ede505c0532f0316ac85c44456c8bf06",
		1000000000,
		56000,
		[]byte("test"),
		integrationTests.ChainID,
		2,
		1,
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
	options uint32,
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
	valMinting.Mul(valMinting, big.NewInt(5000000))

	node := integrationTests.NewTestProcessorNode(
		maxShards,
		nodeShardId,
		txSignPrivKeyShardId,
		initialNodeAddr,
	)

	node.EconomicsData.SetMinGasPrice(10)
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
		Options:   options,
	}
	tx.Value = big.NewInt(0).Set(frontendValue)
	txHexHash, err = node.SendTransaction(tx)
	require.Nil(t, err)

	select {
	case <-chDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout getting transaction")
	}
}
