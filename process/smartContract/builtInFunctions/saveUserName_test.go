package builtInFunctions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestSaveUserName_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	dnsAddr := []byte("DNS")
	mapDnsAddresses := make(map[string]struct{})
	mapDnsAddresses[string(dnsAddr)] = struct{}{}
	coa := saveUserName{
		gasCost:         1,
		mapDnsAddresses: mapDnsAddresses,
		enableChange:    false,
	}

	addr := []byte("addr")

	acc, _ := state.NewUserAccount(mock.NewAddressMock(addr))
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{CallerAddr: dnsAddr, GasProvided: 50},
	}

	_, err := coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, process.ErrInvalidArguments, err)

	_, err = coa.ProcessBuiltinFunction(nil, acc, nil)
	require.Equal(t, process.ErrNilVmInput, err)

	_, err = coa.ProcessBuiltinFunction(nil, nil, vmInput)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	newUserName := []byte("afafafafafafafafafafafafafafafaf")
	vmInput.Arguments = [][]byte{newUserName}
	_, err = coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)

	_, err = coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, process.ErrUserNameChangeIsDisabled, err)
}
