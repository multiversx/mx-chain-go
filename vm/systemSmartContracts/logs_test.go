package systemSmartContracts

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestCreateLogEntryForDelegate(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	delegationValue := big.NewInt(1000)
	res := (&delegation{
		eei: &mock.SystemEIStub{
			GetStorageCalled: func(key []byte) []byte {
				fund := &Fund{
					Value: big.NewInt(5000),
				}
				fundBytes, _ := marshalizer.Marshal(fund)

				return fundBytes
			},
		},
		marshalizer: marshalizer,
	}).createLogEntryForDelegate(
		"identifier",
		[]byte("caller"),
		delegationValue,
		&GlobalFundData{
			TotalActive: big.NewInt(1000000),
		},
		&DelegatorData{
			ActiveFund: []byte("active-fund-key"),
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
