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
	value.SetString("1000000000000000", 10)

	/*
		{
		  nonce: 0,
		  value: '1000000000000000',
		  receiver: 'erd1vpyd9th805tm039nzjfghjd9n9mx7nausl60sz5zf097fuzwurmqkj962a',
		  sender: 'erd1f08fx5gy2jhjuyaw60wemelwl4gnmglert8rdcq7pcgq4m53mdcqf3qa6z',
		  gasPrice: 1000000000,
		  gasLimit: 50000,
		  data: '',
		  chainID: 1,
		  version: 1,
		  signature: 'eadc305b083933489aa1303172c1c0fb79be13110cd1d2612c9a12230f037a96de8dc4c65baef328ef1b0e433ff08ba6fb675c9c708c1f0c2e5f0e1073cb87017b226e6f6e6365223a302c2276616c7565223a2231303030303030303030303030303030222c227265636569766572223a226572643176707964397468383035746d3033396e7a6a6667686a64396e396d78376e6175736c3630737a357a6630393766757a7775726d716b6a39363261222c2273656e646572223a22657264316630386678356779326a686a75796177363077656d656c776c34676e6d676c65727438726463713770636771346d35336d64637166337161367a222c226761735072696365223a313030303030303030302c226761734c696d6974223a35303030302c22636861696e4944223a312c2276657273696f6e223a317d'
		}
	*/
	testInterceptedTxFromFrontendGeneratedParams(
		t,
		0,
		value,
		"erd1vpyd9th805tm039nzjfghjd9n9mx7nausl60sz5zf097fuzwurmqkj962a",
		"erd1f08fx5gy2jhjuyaw60wemelwl4gnmglert8rdcq7pcgq4m53mdcqf3qa6z",
		"eadc305b083933489aa1303172c1c0fb79be13110cd1d2612c9a12230f037a96de8dc4c65baef328ef1b0e433ff08ba6fb675c9c708c1f0c2e5f0e1073cb87017b226e6f6e6365223a302c2276616c7565223a2231303030303030303030303030303030222c227265636569766572223a226572643176707964397468383035746d3033396e7a6a6667686a64396e396d78376e6175736c3630737a357a6630393766757a7775726d716b6a39363261222c2273656e646572223a22657264316630386678356779326a686a75796177363077656d656c776c34676e6d676c65727438726463713770636771346d35336d64637166337161367a222c226761735072696365223a313030303030303030302c226761734c696d6974223a35303030302c22636861696e4944223a312c2276657273696f6e223a317d",
		1000000000,
		50000,
		[]byte(""),
		[]byte("1"),
		1,
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
