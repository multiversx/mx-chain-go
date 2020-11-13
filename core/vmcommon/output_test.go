package vmcommon

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFirstReturnData_VMOutputWithNoReturnDataShouldErr(t *testing.T) {
	vmOutput := VMOutput{
		ReturnData: [][]byte{},
	}

	_, err := vmOutput.GetFirstReturnData(AsBigInt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no return data")
}

func TestGetFirstReturnData_WithBadReturnDataKindShouldErr(t *testing.T) {
	vmOutput := VMOutput{
		ReturnData: [][]byte{[]byte("100")},
	}

	_, err := vmOutput.GetFirstReturnData(42)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "can't interpret")
}

func TestGetFirstReturnData(t *testing.T) {
	value := big.NewInt(100)

	vmOutput := VMOutput{
		ReturnData: [][]byte{value.Bytes()},
	}

	dataAsBigInt, _ := vmOutput.GetFirstReturnData(AsBigInt)
	dataAsBigIntString, _ := vmOutput.GetFirstReturnData(AsBigIntString)
	dataAsString, _ := vmOutput.GetFirstReturnData(AsString)
	dataAsHex, _ := vmOutput.GetFirstReturnData(AsHex)

	assert.Equal(t, value, dataAsBigInt)
	assert.Equal(t, "100", dataAsBigIntString)
	assert.Equal(t, string(value.Bytes()), dataAsString)
	assert.Equal(t, "64", dataAsHex)
}
