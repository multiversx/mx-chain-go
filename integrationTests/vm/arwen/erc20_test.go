package arwen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SOL_002(t *testing.T) {
	context := setupTestContext(t)

	context.deploySC("./testdata/erc20sol002/0-0-2.wasm", "")

	// Initial tokens and allowances
	context.executeSC(&context.Owner, "transfer(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(&context.Owner, "transfer(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(1000))

	context.executeSC(&context.Alice, "increaseAllowance(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(&context.Alice, "decreaseAllowance(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(5))
	context.executeSC(&context.Bob, "increaseAllowance(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(&context.Bob, "decreaseAllowance(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(5))

	// Assertion
	assert.Equal(t, uint64(42000), context.querySCInt("totalSupply()", [][]byte{}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{context.Bob.Address}))
	assert.Equal(t, uint64(495), context.querySCInt("allowance(address,address)", [][]byte{context.Alice.Address, context.Bob.Address}))
	assert.Equal(t, uint64(495), context.querySCInt("allowance(address,address)", [][]byte{context.Bob.Address, context.Alice.Address}))

	// Payments
	context.executeSC(&context.Alice, "transferFrom(address,address,uint256)@"+context.Bob.AddressHex()+"@"+context.Carol.AddressHex()+"@"+formatHexNumber(50))
	context.executeSC(&context.Bob, "transferFrom(address,address,uint256)@"+context.Alice.AddressHex()+"@"+context.Carol.AddressHex()+"@"+formatHexNumber(50))

	// Assertion
	assert.Equal(t, uint64(950), context.querySCInt("balanceOf(address)", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(950), context.querySCInt("balanceOf(address)", [][]byte{context.Bob.Address}))
	assert.Equal(t, uint64(100), context.querySCInt("balanceOf(address)", [][]byte{context.Carol.Address}))
}

func Test_SOL_003(t *testing.T) {
	context := setupTestContext(t)

	context.deploySC("./testdata/erc20sol003/0-0-3.wasm", "")

	// Minting
	context.executeSC(&context.Owner, "transfer(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(&context.Owner, "transfer(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(1000))

	// Assertion
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{context.Bob.Address}))

	// Regular transfers
	context.executeSC(&context.Alice, "transfer(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(200))
	context.executeSC(&context.Bob, "transfer(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(400))

	// Assertion
	assert.Equal(t, uint64(1200), context.querySCInt("balanceOf(address)", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(800), context.querySCInt("balanceOf(address)", [][]byte{context.Bob.Address}))
}

func Test_C_001(t *testing.T) {
	context := setupTestContext(t)

	context.deploySC("./testdata/erc20c/wrc20_arwen.wasm", formatHexNumber(42000))

	// Assertion
	assert.Equal(t, uint64(42000), context.querySCInt("totalSupply", [][]byte{}))
	assert.Equal(t, uint64(42000), context.querySCInt("balanceOf", [][]byte{context.Owner.Address}))

	// Minting
	context.executeSC(&context.Owner, "transferToken@"+context.Alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(&context.Owner, "transferToken@"+context.Bob.AddressHex()+"@"+formatHexNumber(1000))

	// Regular transfers
	context.executeSC(&context.Alice, "transferToken@"+context.Bob.AddressHex()+"@"+formatHexNumber(200))
	context.executeSC(&context.Bob, "transferToken@"+context.Alice.AddressHex()+"@"+formatHexNumber(400))

	// Assertion
	assert.Equal(t, uint64(1200), context.querySCInt("balanceOf", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(800), context.querySCInt("balanceOf", [][]byte{context.Bob.Address}))

	// Approve and transfer
	context.executeSC(&context.Alice, "approve@"+context.Bob.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(&context.Bob, "approve@"+context.Alice.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(&context.Alice, "transferFrom@"+context.Bob.AddressHex()+"@"+context.Carol.AddressHex()+"@"+formatHexNumber(25))
	context.executeSC(&context.Bob, "transferFrom@"+context.Alice.AddressHex()+"@"+context.Carol.AddressHex()+"@"+formatHexNumber(25))

	assert.Equal(t, uint64(1175), context.querySCInt("balanceOf", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(775), context.querySCInt("balanceOf", [][]byte{context.Bob.Address}))
	assert.Equal(t, uint64(50), context.querySCInt("balanceOf", [][]byte{context.Carol.Address}))
}
