package smartcontract

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/stretchr/testify/require"
)

func TestDNS_Register(t *testing.T) {
	// TODO reinstate test after Wasm VM pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	expectedDNSAddress := []byte{0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 180, 108, 178, 102, 195, 67, 184, 127, 204, 159, 104, 123, 190, 33, 224, 91, 255, 244, 118, 95, 24, 217}

	var empty struct{}
	wasm.DNSAddresses[string(expectedDNSAddress)] = empty
	gasScheduleConfigPath := strings.ReplaceAll(integrationTests.GasSchedulePath, "../../../..", "../../..")

	context := wasm.SetupTestContextWithGasSchedulePath(t, gasScheduleConfigPath)
	defer context.Close()

	context.GasLimit = 40000000
	err := context.DeploySC("dns.wasm", "0064")
	require.Nil(t, err)
	require.True(t, bytes.Equal(expectedDNSAddress, context.ScAddress))

	name := "thisisalice398"
	testname := hex.EncodeToString([]byte(name))
	context.GasLimit = 80000000
	err = context.ExecuteSCWithValue(&context.Alice, "register@"+testname, big.NewInt(100))
	require.Nil(t, err)

	context.GasLimit = 13000000
	err = context.ExecuteSCWithValue(&context.Alice, "resolve@"+testname, big.NewInt(0))
	require.Nil(t, err)

	for _, scr := range context.LastSCResults {
		if bytes.Equal(scr.GetOriginalTxHash(), context.LastTxHash) {
			data := scr.GetData()
			if len(data) > 0 {
				// The first 6 characters of data are '@6f6b@', where 6f6b means 'ok';
				// the resolved address comes after.
				resolvedAddress, err := hex.DecodeString(string(data[6:]))
				require.Nil(t, err)
				require.True(t, bytes.Equal(context.Alice.Address, resolvedAddress))
			}
		}
	}
}
