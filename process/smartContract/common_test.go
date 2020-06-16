package smartContract

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func Test_parseVMTypeFromContractAddress(t *testing.T) {
	fakeVMType, _ := hex.DecodeString("beaf")
	scAddress, _ := hex.DecodeString("0000000000000000beaf00000000000022cd8429ce92f8973bba2a9fb51e0eb3a1")
	vmType, err := parseVMTypeFromContractAddress(scAddress)
	require.Nil(t, err)
	require.Equal(t, fakeVMType, vmType)
}

func Test_parseVMTypeFromContractAddress_WhenAddressTooShort(t *testing.T) {
	scAddress, _ := hex.DecodeString("0003a1")
	vmType, err := parseVMTypeFromContractAddress(scAddress)
	require.Error(t, err, process.ErrInvalidVMType)
	require.Nil(t, vmType)
}
