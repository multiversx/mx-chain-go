package sctests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SOL_002(t *testing.T) {
	context := setupTestContext(t)

	context.deploySC("./erc20sol002/0-0-2.wasm", "")

	// Initial tokens and allowances
	context.executeSC(&context.Owner, "transfer(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(&context.Owner, "transfer(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(1000))

	context.executeSC(&context.Owner, "increaseAllowance(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(&context.Owner, "increaseAllowance(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(500))
	context.executeSC(&context.Owner, "decreaseAllowance(address,uint256)@"+context.Alice.AddressHex()+"@"+formatHexNumber(5))
	context.executeSC(&context.Owner, "decreaseAllowance(address,uint256)@"+context.Bob.AddressHex()+"@"+formatHexNumber(5))

	// Assertion
	assert.Equal(t, uint64(42000), context.querySCInt("totalSupply()", [][]byte{}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(1000), context.querySCInt("balanceOf(address)", [][]byte{context.Bob.Address}))
	assert.Equal(t, uint64(495), context.querySCInt("allowance(address,address)", [][]byte{context.Owner.Address, context.Alice.Address}))
	assert.Equal(t, uint64(495), context.querySCInt("allowance(address,address)", [][]byte{context.Owner.Address, context.Bob.Address}))

	// Payments
	context.executeSC(&context.Alice, "transferFrom(address,address,uint256)@"+context.Alice.AddressHex()+"@"+context.Bob.AddressHex()+"@"+formatHexNumber(5))
	context.executeSC(&context.Alice, "transferFrom(address,address,uint256)@"+context.Alice.AddressHex()+"@"+context.Bob.AddressHex()+"@"+formatHexNumber(5))

	// Assertion
	// TODO: Why doesn't balance change?
	// TODO: Add some meaningful assertions.
}

func Test_SOL_003(t *testing.T) {
	context := setupTestContext(t)

	context.deploySC("./erc20sol003/0-0-3.wasm", "")

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

	context.deploySC("./erc20c/wrc20_arwen.wasm", formatHexNumber(42000))

	// Assertion
	// TODO: Check why totalSupply isn't 42000 (following assertions fail).
	// assert.Equal(t, uint64(42000), context.querySCInt("totalSupply", [][]byte{}))
	//assert.Equal(t, uint64(42000), context.querySCInt("balanceOf", [][]byte{context.Owner.Address}))

	// Minting
	context.executeSC(&context.Owner, "transferToken@"+context.Alice.AddressHex()+"@"+formatHexNumber(1000))
	context.executeSC(&context.Owner, "transferToken@"+context.Bob.AddressHex()+"@"+formatHexNumber(1000))

	// Regular transfers
	context.executeSC(&context.Alice, "transferToken@"+context.Bob.AddressHex()+"@00c8") // because SC sees "c8" as a negative number and signals an error
	context.executeSC(&context.Bob, "transferToken@"+context.Alice.AddressHex()+"@"+formatHexNumber(400))

	// Assertion
	assert.Equal(t, uint64(1200), context.querySCInt("balanceOf", [][]byte{context.Alice.Address}))
	assert.Equal(t, uint64(800), context.querySCInt("balanceOf", [][]byte{context.Bob.Address}))
}
