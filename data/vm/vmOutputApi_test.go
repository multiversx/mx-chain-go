package vm

import (
	"encoding/hex"
	"math/big"
	"testing"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestVMOutputApi_GetFirstReturnDataAsBigInt(t *testing.T) {
	t.Parallel()

	expectedRes := 37

	vmOutputApi := VMOutputApi{
		ReturnData: [][]byte{big.NewInt(int64(expectedRes)).Bytes()},
	}

	res, err := vmOutputApi.GetFirstReturnData(vmcommon.AsBigInt)
	require.NoError(t, err)

	resBigInt, ok := res.(*big.Int)
	require.True(t, ok)
	require.Equal(t, uint64(expectedRes), resBigInt.Uint64())
}

func TestVMOutputApi_GetFirstReturnDataAsBigIntString(t *testing.T) {
	t.Parallel()

	expectedRes := "37"

	bi, _ := big.NewInt(0).SetString(expectedRes, 10)
	vmOutputApi := VMOutputApi{
		ReturnData: [][]byte{bi.Bytes()},
	}

	res, err := vmOutputApi.GetFirstReturnData(vmcommon.AsBigIntString)
	require.NoError(t, err)

	resBigIntString, ok := res.(string)
	require.True(t, ok)
	require.Equal(t, expectedRes, resBigIntString)
}

func TestVMOutputApi_GetFirstReturnDataAsHex(t *testing.T) {
	t.Parallel()

	expectedRes := hex.EncodeToString([]byte("37"))

	resBytes, _ := hex.DecodeString(expectedRes)
	vmOutputApi := VMOutputApi{
		ReturnData: [][]byte{resBytes},
	}

	res, err := vmOutputApi.GetFirstReturnData(vmcommon.AsHex)
	require.NoError(t, err)

	resHexString, ok := res.(string)
	require.True(t, ok)
	require.Equal(t, expectedRes, resHexString)
}

func TestVMOutputApi_GetFirstReturnDataAsString(t *testing.T) {
	t.Parallel()

	expectedRes := "37"
	vmOutputApi := VMOutputApi{
		ReturnData: [][]byte{[]byte(expectedRes)},
	}

	res, err := vmOutputApi.GetFirstReturnData(vmcommon.AsString)
	require.NoError(t, err)

	resString, ok := res.(string)
	require.True(t, ok)
	require.Equal(t, expectedRes, resString)
}
