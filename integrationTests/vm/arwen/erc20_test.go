package arwen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SOL_002(t *testing.T) {
	context := setupTestContext(t)

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob
	carol := &context.Carol

	context.deploySC("./testdata/erc20/0-0-2.wasm", "")

	// Initial tokens and allowances
	context.executeSC(owner, "transfer(address,uint256)@"+alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(owner, "transfer(address,uint256)@"+bob.AddressHex()+"@"+formatHexNumber(1000))

	context.executeSC(alice, "increaseAllowance(address,uint256)@"+bob.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(alice, "decreaseAllowance(address,uint256)@"+bob.AddressHex()+"@"+formatHexNumber(5))
	context.executeSC(bob, "increaseAllowance(address,uint256)@"+alice.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(bob, "decreaseAllowance(address,uint256)@"+alice.AddressHex()+"@"+formatHexNumber(5))

	// Assertion
	assert.Equal(t, uint64(42000), context.querySCInt("totalSupply()", [][]byte{}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{bob.Address}))
	assert.Equal(t, uint64(495), context.querySCInt("allowance(address,address)", [][]byte{alice.Address, bob.Address}))
	assert.Equal(t, uint64(495), context.querySCInt("allowance(address,address)", [][]byte{bob.Address, alice.Address}))

	// Payments
	context.executeSC(alice, "transferFrom(address,address,uint256)@"+bob.AddressHex()+"@"+carol.AddressHex()+"@"+formatHexNumber(50))
	context.executeSC(bob, "transferFrom(address,address,uint256)@"+alice.AddressHex()+"@"+carol.AddressHex()+"@"+formatHexNumber(50))

	// Assertion
	assert.Equal(t, uint64(950), context.querySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(950), context.querySCInt("balanceOf(address)", [][]byte{bob.Address}))
	assert.Equal(t, uint64(100), context.querySCInt("balanceOf(address)", [][]byte{carol.Address}))
}

func Test_SOL_003(t *testing.T) {
	context := setupTestContext(t)

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob

	context.deploySC("./testdata/erc20/0-0-3.wasm", "")

	// Minting
	context.executeSC(owner, "transfer(address,uint256)@"+alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(owner, "transfer(address,uint256)@"+bob.AddressHex()+"@"+formatHexNumber(1000))

	// Assertion
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{bob.Address}))

	// Regular transfers
	context.executeSC(alice, "transfer(address,uint256)@"+bob.AddressHex()+"@"+formatHexNumber(200))
	context.executeSC(bob, "transfer(address,uint256)@"+alice.AddressHex()+"@"+formatHexNumber(400))

	// Assertion
	assert.Equal(t, uint64(1200), context.querySCInt("balanceOf(address)", [][]byte{alice.Address}))
	assert.Equal(t, uint64(800), context.querySCInt("balanceOf(address)", [][]byte{bob.Address}))
}

func Test_C_001(t *testing.T) {
	context := setupTestContext(t)

	owner := &context.Owner
	alice := &context.Alice
	bob := &context.Bob
	carol := &context.Carol

	context.deploySC("./testdata/erc20/wrc20_arwen_03.wasm", formatHexNumber(42000))

	// Assertion
	assert.Equal(t, uint64(42000), context.querySCInt("totalSupply", [][]byte{}))
	assert.Equal(t, uint64(42000), context.querySCInt("balanceOf", [][]byte{context.Owner.Address}))

	// Minting
	context.executeSC(owner, "transferToken@"+alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(owner, "transferToken@"+bob.AddressHex()+"@"+formatHexNumber(1000))

	// Regular transfers
	context.executeSC(alice, "transferToken@"+bob.AddressHex()+"@"+formatHexNumber(200))
	context.executeSC(bob, "transferToken@"+alice.AddressHex()+"@"+formatHexNumber(400))

	// Assertion
	assert.Equal(t, uint64(1200), context.querySCInt("balanceOf", [][]byte{alice.Address}))
	assert.Equal(t, uint64(800), context.querySCInt("balanceOf", [][]byte{bob.Address}))

	// Approve and transfer
	context.executeSC(alice, "approve@"+bob.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(bob, "approve@"+alice.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(alice, "transferFrom@"+bob.AddressHex()+"@"+carol.AddressHex()+"@"+formatHexNumber(25))
	context.executeSC(bob, "transferFrom@"+alice.AddressHex()+"@"+carol.AddressHex()+"@"+formatHexNumber(25))

	assert.Equal(t, uint64(1175), context.querySCInt("balanceOf", [][]byte{alice.Address}))
	assert.Equal(t, uint64(775), context.querySCInt("balanceOf", [][]byte{bob.Address}))
	assert.Equal(t, uint64(50), context.querySCInt("balanceOf", [][]byte{carol.Address}))
}
