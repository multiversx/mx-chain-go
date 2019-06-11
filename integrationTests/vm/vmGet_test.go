package vm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVmGetShouldReturnValue(t *testing.T) {
	opGas := uint64(0)
	transferValue := big.NewInt(0)

	accnts, destinationAddressBytes, expectedValueForVar := deployAndRunSmartContract(t, opGas, transferValue)

	scgd := createScDataGetterWithOneSCExecutorMockVM(accnts, opGas)
	functionName := "Get"
	returnedVals, err := scgd.Get(destinationAddressBytes, functionName)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(returnedVals))
	assert.Equal(t, expectedValueForVar.Bytes(), returnedVals)
}
