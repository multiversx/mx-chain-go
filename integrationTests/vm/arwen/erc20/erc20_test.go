package erc20

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/assert"
)

func Test_SOL_002(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob
	carol := &context.Carol

	context.DeploySC("../testdata/erc20-solidity-002/0-0-2.wasm", "")

	// Initial tokens and allowances
	context.ExecuteSC(owner, "transfer(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(1000))
	context.ExecuteSC(owner, "transfer(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(1000))

	context.ExecuteSC(alice, "increaseAllowance(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(500))
	context.ExecuteSC(alice, "decreaseAllowance(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(5))
	context.ExecuteSC(bob, "increaseAllowance(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(500))
	context.ExecuteSC(bob, "decreaseAllowance(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(5))

	// Assertion
	assert.Equal(t, uint64(42000), context.QuerySCInt("totalSupply()", [][]byte{}))
	assert.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))
	assert.Equal(t, uint64(495), context.QuerySCInt("allowance(address,address)", [][]byte{alice.Address, bob.Address}))
	assert.Equal(t, uint64(495), context.QuerySCInt("allowance(address,address)", [][]byte{bob.Address, alice.Address}))

	// Payments
	context.ExecuteSC(alice, "transferFrom(address,address,uint256)@"+bob.AddressHex()+"@"+carol.AddressHex()+"@"+arwen.FormatHexNumber(50))
	context.ExecuteSC(bob, "transferFrom(address,address,uint256)@"+alice.AddressHex()+"@"+carol.AddressHex()+"@"+arwen.FormatHexNumber(50))

	// Assertion
	assert.Equal(t, uint64(950), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(950), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))
	assert.Equal(t, uint64(100), context.QuerySCInt("balanceOf(address)", [][]byte{carol.Address}))
}

func Test_SOL_003(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob

	context.DeploySC("../testdata/erc20-solidity-003/0-0-3.wasm", "")

	// Minting
	context.ExecuteSC(owner, "transfer(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(1000))
	context.ExecuteSC(owner, "transfer(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(1000))

	// Assertion
	assert.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(1000), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))

	// Regular transfers
	context.ExecuteSC(alice, "transfer(address,uint256)@"+bob.AddressHex()+"@"+arwen.FormatHexNumber(200))
	context.ExecuteSC(bob, "transfer(address,uint256)@"+alice.AddressHex()+"@"+arwen.FormatHexNumber(400))

	// Assertion
	assert.Equal(t, uint64(1200), context.QuerySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(800), context.QuerySCInt("balanceOf(address)", [][]byte{bob.Address}))
}

func Test_C_001(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob
	carol := &context.Carol

	context.DeploySC("../testdata/erc20-c-03/wrc20_arwen.wasm", "00"+arwen.FormatHexNumber(42000))

	// Assertion
	assert.Equal(t, uint64(42000), context.QuerySCInt("totalSupply", [][]byte{}))
	assert.Equal(t, uint64(42000), context.QuerySCInt("balanceOf", [][]byte{context.Owner.Address}))

	// Minting
	context.ExecuteSC(owner, "transferToken@"+alice.AddressHex()+"@00"+arwen.FormatHexNumber(1000))
	context.ExecuteSC(owner, "transferToken@"+bob.AddressHex()+"@00"+arwen.FormatHexNumber(1000))

	// Regular transfers
	context.ExecuteSC(alice, "transferToken@"+bob.AddressHex()+"@00"+arwen.FormatHexNumber(200))
	context.ExecuteSC(bob, "transferToken@"+alice.AddressHex()+"@00"+arwen.FormatHexNumber(400))

	// Assertion
	assert.Equal(t, uint64(1200), context.QuerySCInt("balanceOf", [][]byte{alice.Address}))
	assert.Equal(t, uint64(800), context.QuerySCInt("balanceOf", [][]byte{bob.Address}))

	// Approve and transfer
	context.ExecuteSC(alice, "approve@"+bob.AddressHex()+"@00"+arwen.FormatHexNumber(500))
	context.ExecuteSC(bob, "approve@"+alice.AddressHex()+"@00"+arwen.FormatHexNumber(500))
	context.ExecuteSC(alice, "transferFrom@"+bob.AddressHex()+"@"+carol.AddressHex()+"@00"+arwen.FormatHexNumber(25))
	context.ExecuteSC(bob, "transferFrom@"+alice.AddressHex()+"@"+carol.AddressHex()+"@00"+arwen.FormatHexNumber(25))

	assert.Equal(t, uint64(1175), context.QuerySCInt("balanceOf", [][]byte{alice.Address}))
	assert.Equal(t, uint64(775), context.QuerySCInt("balanceOf", [][]byte{bob.Address}))
	assert.Equal(t, uint64(50), context.QuerySCInt("balanceOf", [][]byte{carol.Address}))
}
