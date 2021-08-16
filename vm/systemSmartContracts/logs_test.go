package systemSmartContracts

import (
	"math/big"
	"testing"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestCreateLogEntryForDelegate(t *testing.T) {
	t.Parallel()

	delegationValue := big.NewInt(1000)
	res := createLogEntryForDelegate(
		"identifier",
		[]byte("caller"),
		delegationValue,
		&GlobalFundData{
			TotalActive: big.NewInt(1000000),
		},
		&DelegatorData{
			ActiveFund: big.NewInt(5000).Bytes(),
		},
		&DelegationContractStatus{},
		true,
	)

	require.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte("identifier"),
		Address:    []byte("caller"),
		Topics:     [][]byte{delegationValue.Bytes(), big.NewInt(6000).Bytes(), big.NewInt(1).Bytes(), big.NewInt(1001000).Bytes()},
	}, res)
}
