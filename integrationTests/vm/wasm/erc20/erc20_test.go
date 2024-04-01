package erc20

import (
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/stretchr/testify/require"
)

func Test_C_001(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	context := wasm.SetupTestContext(t)
	defer context.Close()

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob
	carol := &context.Carol

	err := context.DeploySC("../testdata/erc20-c-03/wrc20_wasm.wasm", "00"+wasm.FormatHexNumber(42000))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(42000), context.QuerySCInt("totalSupply", [][]byte{}))
	require.Equal(t, uint64(42000), context.QuerySCInt("balanceOf", [][]byte{context.Owner.Address}))

	// Minting
	err = context.ExecuteSC(owner, "transferToken@"+alice.AddressHex()+"@00"+wasm.FormatHexNumber(1000))
	require.Nil(t, err)
	err = context.ExecuteSC(owner, "transferToken@"+bob.AddressHex()+"@00"+wasm.FormatHexNumber(1000))
	require.Nil(t, err)

	// Regular transfers
	err = context.ExecuteSC(alice, "transferToken@"+bob.AddressHex()+"@00"+wasm.FormatHexNumber(200))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "transferToken@"+alice.AddressHex()+"@00"+wasm.FormatHexNumber(400))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(1200), context.QuerySCInt("balanceOf", [][]byte{alice.Address}))
	require.Equal(t, uint64(800), context.QuerySCInt("balanceOf", [][]byte{bob.Address}))

	// Approve and transfer
	err = context.ExecuteSC(alice, "approve@"+bob.AddressHex()+"@00"+wasm.FormatHexNumber(500))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "approve@"+alice.AddressHex()+"@00"+wasm.FormatHexNumber(500))
	require.Nil(t, err)
	err = context.ExecuteSC(alice, "transferFrom@"+bob.AddressHex()+"@"+carol.AddressHex()+"@00"+wasm.FormatHexNumber(25))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "transferFrom@"+alice.AddressHex()+"@"+carol.AddressHex()+"@00"+wasm.FormatHexNumber(25))
	require.Nil(t, err)

	require.Equal(t, uint64(1175), context.QuerySCInt("balanceOf", [][]byte{alice.Address}))
	require.Equal(t, uint64(775), context.QuerySCInt("balanceOf", [][]byte{bob.Address}))
	require.Equal(t, uint64(50), context.QuerySCInt("balanceOf", [][]byte{carol.Address}))
}
