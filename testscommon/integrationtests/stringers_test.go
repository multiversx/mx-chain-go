package integrationtests

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

const addressSize = 32
const hashSize = 32

var pkConv = testscommon.RealWorldBech32PubkeyConverter

func TestSmartContractResultsToString(t *testing.T) {
	t.Parallel()

	scr1 := &smartContractResult.SmartContractResult{
		Value: big.NewInt(0),
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          1,
		Value:          big.NewInt(2),
		RcvAddr:        bytes.Repeat([]byte{0}, addressSize),
		SndAddr:        bytes.Repeat([]byte{1}, addressSize),
		RelayerAddr:    bytes.Repeat([]byte{2}, addressSize),
		RelayedValue:   big.NewInt(3),
		Code:           []byte("code"),
		Data:           []byte("data"),
		PrevTxHash:     bytes.Repeat([]byte{3}, hashSize),
		OriginalTxHash: bytes.Repeat([]byte{4}, hashSize),
		GasLimit:       4,
		GasPrice:       5,
		CallType:       vm.AsynchronousCall,
		CodeMetadata:   []byte{1, 0},
		ReturnMessage:  []byte("return message"),
		OriginalSender: bytes.Repeat([]byte{6}, addressSize),
	}

	result := SmartContractResultsToString(pkConv, scr1, scr2, nil)
	expected := `[
	SmartContractResult %p{
		Nonce: 0
		Value: 0
		RcvAddr: 
		SndAddr: 
		OriginalSender: 
		ReturnMessage: 
		Data: 
		CallType: directCall
		Code: 
		GasLimit: 0
		GasPrice: 0
		OriginalTxHash: 
		PrevTxHash: 
		RelayerAddr: 
		RelayedValue: 0
	}
	SmartContractResult %p{
		Nonce: 1
		Value: 2
		RcvAddr: erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu
		SndAddr: erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7
		OriginalSender: erd1qcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqwkh39e
		ReturnMessage: return message
		Data: data
		CallType: asynchronousCall
		Code: code
		GasLimit: 4
		GasPrice: 5
		OriginalTxHash: 0404040404040404040404040404040404040404040404040404040404040404
		PrevTxHash: 0303030303030303030303030303030303030303030303030303030303030303
		RelayerAddr: erd1qgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqjv7g5k
		RelayedValue: 3
	}
	SmartContractResult !NIL{}
]`
	assert.Equal(t, fmt.Sprintf(expected, scr1, scr2), result)
}
