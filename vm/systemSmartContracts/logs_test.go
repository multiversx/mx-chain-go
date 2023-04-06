package systemSmartContracts

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestCreateLogEntryForDelegate(t *testing.T) {
	t.Parallel()

	var res *vmcommon.LogEntry
	marshalizer := &mock.MarshalizerMock{}
	delegationValue := big.NewInt(1000)
	(&delegation{
		eei: &mock.SystemEIStub{
			GetStorageCalled: func(key []byte) []byte {
				fund := &Fund{
					Value: big.NewInt(5000),
				}
				fundBytes, _ := marshalizer.Marshal(fund)

				return fundBytes
			},
			AddLogEntryCalled: func(entry *vmcommon.LogEntry) {
				res = entry
			},
		},
		marshalizer: marshalizer,
	}).createAndAddLogEntryForDelegate(
		&vmcommon.ContractCallInput{
			Function: "identifier",
			VMInput: vmcommon.VMInput{
				CallerAddr: []byte("caller"),
			},
		},
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
		Identifier: []byte("delegate"),
		Address:    []byte("caller"),
		Topics:     [][]byte{delegationValue.Bytes(), big.NewInt(6000).Bytes(), big.NewInt(1).Bytes(), big.NewInt(1001000).Bytes()},
	}, res)
}

func TestCreateAndAddLogEntryForWithdraw(t *testing.T) {
	t.Parallel()

	var res *vmcommon.LogEntry
	marshalizer := &mock.MarshalizerMock{}
	actualUserUnBond := big.NewInt(1000)
	(&delegation{
		eei: &mock.SystemEIStub{
			GetStorageCalled: func(key []byte) []byte {
				fund := &Fund{
					Value: big.NewInt(5000),
				}
				fundBytes, _ := marshalizer.Marshal(fund)

				return fundBytes
			},
			AddLogEntryCalled: func(entry *vmcommon.LogEntry) {
				res = entry
			},
		},
		marshalizer: marshalizer,
	}).createAndAddLogEntryForWithdraw(
		"withdraw",
		[]byte("caller"),
		actualUserUnBond,
		&GlobalFundData{
			TotalActive: big.NewInt(1000000),
		},
		&DelegatorData{
			ActiveFund: []byte("active-fund-key"),
		},
		1,
		false,
		[][]byte{[]byte("fund-id-1"), []byte("fund-id-2")},
	)

	require.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte("withdraw"),
		Address:    []byte("caller"),
		Topics:     [][]byte{actualUserUnBond.Bytes(), big.NewInt(5000).Bytes(), big.NewInt(1).Bytes(), big.NewInt(1000000).Bytes(), []byte(strconv.FormatBool(false)), []byte("fund-id-1"), []byte("fund-id-2")},
	}, res)
}
