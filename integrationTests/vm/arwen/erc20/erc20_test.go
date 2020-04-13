package erc20

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/require"
)

func Test_SOL_002(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob
	carol := &context.Carol

	err := context.DeploySC("../testdata/erc20-solidity-002/0-0-2.wasm", "")
	require.Nil(t, err)

	// Initial tokens and allowances
	err = context.ExecuteSC(owner, "transfer(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(1000))
	require.Nil(t, err)
	err = context.ExecuteSC(owner, "transfer(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(1000))
	require.Nil(t, err)

	err = context.ExecuteSC(alice, "increaseAllowance(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(500))
	require.Nil(t, err)
	err = context.ExecuteSC(alice, "decreaseAllowance(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(5))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "increaseAllowance(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(500))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "decreaseAllowance(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(5))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(42000), context.QuerySCInt("totalSupply()", [][]byte{}))
	require.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	require.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))
	require.Equal(t, uint64(495), context.QuerySCInt("allowance(address,address)", [][]byte{alice.Address, bob.Address}))
	require.Equal(t, uint64(495), context.QuerySCInt("allowance(address,address)", [][]byte{bob.Address, alice.Address}))

	// Payments
	err = context.ExecuteSC(alice, "transferFrom(address,address,uint256)@"+bob.AddressHex()+"@"+carol.AddressHex()+"@"+arwen.FormatHexNumber(50))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "transferFrom(address,address,uint256)@"+alice.AddressHex()+"@"+carol.AddressHex()+"@"+arwen.FormatHexNumber(50))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(950), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	require.Equal(t, uint64(950), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))
	require.Equal(t, uint64(100), context.QuerySCInt("balanceOf(address)", [][]byte{carol.Address}))
}

func Test_SOL_003(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob

	err := context.DeploySC("../testdata/erc20-solidity-003/0-0-3.wasm", "")
	require.Nil(t, err)

	// Minting
	err = context.ExecuteSC(owner, "transfer(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(1000))
	require.Nil(t, err)
	err = context.ExecuteSC(owner, "transfer(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(1000))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	require.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))

	// Regular transfers
	err = context.ExecuteSC(alice, "transfer(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(200))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "transfer(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(400))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(1200), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	require.Equal(t, uint64(800), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))
}

func Test_C_001(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob
	carol := &context.Carol

	err := context.DeploySC("../testdata/erc20-c-03/wrc20_arwen.wasm", "00"+arwen.FormatHexNumber(42000))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(42000), context.QuerySCInt("totalSupply", [][]byte{}))
	require.Equal(t, uint64(42000), context.QuerySCInt("balanceOf", [][]byte{context.Owner.Address}))

	// Minting
	err = context.ExecuteSC(owner, "transferToken@"+alice.AddressHex()+"@00"+arwen.FormatHexNumber(1000))
	require.Nil(t, err)
	err = context.ExecuteSC(owner, "transferToken@"+bob.AddressHex()+"@00"+arwen.FormatHexNumber(1000))
	require.Nil(t, err)

	// Regular transfers
	err = context.ExecuteSC(alice, "transferToken@"+bob.AddressHex()+"@00"+arwen.FormatHexNumber(200))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "transferToken@"+alice.AddressHex()+"@00"+arwen.FormatHexNumber(400))
	require.Nil(t, err)

	// Assertion
	require.Equal(t, uint64(1200), context.QuerySCInt("balanceOf", [][]byte{alice.Address}))
	require.Equal(t, uint64(800), context.QuerySCInt("balanceOf", [][]byte{bob.Address}))

	// Approve and transfer
	err = context.ExecuteSC(alice, "approve@"+bob.AddressHex()+"@00"+arwen.FormatHexNumber(500))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "approve@"+alice.AddressHex()+"@00"+arwen.FormatHexNumber(500))
	require.Nil(t, err)
	err = context.ExecuteSC(alice, "transferFrom@"+bob.AddressHex()+"@"+carol.AddressHex()+"@00"+arwen.FormatHexNumber(25))
	require.Nil(t, err)
	err = context.ExecuteSC(bob, "transferFrom@"+alice.AddressHex()+"@"+carol.AddressHex()+"@00"+arwen.FormatHexNumber(25))
	require.Nil(t, err)

	require.Equal(t, uint64(1175), context.QuerySCInt("balanceOf", [][]byte{alice.Address}))
	require.Equal(t, uint64(775), context.QuerySCInt("balanceOf", [][]byte{bob.Address}))
	require.Equal(t, uint64(50), context.QuerySCInt("balanceOf", [][]byte{carol.Address}))
}
