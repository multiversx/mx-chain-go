package attestation

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestDNS_Register(t *testing.T) {
	t.Skip("Skip until Arwen update")

	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.ScCodeMetadata.Upgradeable = true
	context.GasLimit = 40000000

	err := context.DeploySC("../testdata/dns/dns.wasm", "0064")
	require.Nil(t, err)

	vmOutput, err := context.QueryService.ExecuteQuery(&process.SCQuery{
		ScAddress: context.ScAddress,
		FuncName:  "register",
		Arguments: [][]byte{[]byte("6794ba4f8dbf57fb0b13")},
		CallValue: big.NewInt(100),
	})

	require.Nil(t, err)
	require.NotNil(t, vmOutput)
}
