package txDataBuilder

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

// TxDataBuilder constructs a string to be used for transaction arguments
type TxDataBuilder struct {
	function  string
	elements  []string
	separator string
}

// NewBuilder creates a new txDataBuilder instance.
func NewBuilder() *TxDataBuilder {
	return &TxDataBuilder{
		function:  "",
		elements:  make([]string, 0),
		separator: "@",
	}
}

// Clear resets the internal state of the txDataBuilder, allowing a new data
// string to be built.
func (builder *TxDataBuilder) Clear() *TxDataBuilder {
	builder.function = ""
	builder.elements = make([]string, 0)

	return builder
}

// Elements returns the individual elements added to the builder
func (builder *TxDataBuilder) Elements() []string {
	return builder.elements
}

// ToString returns the data as a string.
func (builder *TxDataBuilder) ToString() string {
	if len(builder.function) > 0 {
		return builder.toStringWithFunction()
	}

	return builder.toStringWithoutFunction()
}

// ToBytes returns the data as a slice of bytes.
func (builder *TxDataBuilder) ToBytes() []byte {
	return []byte(builder.ToString())
}

// GetLast returns the currently last element.
func (builder *TxDataBuilder) GetLast() string {
	if len(builder.elements) == 0 {
		return ""
	}

	return builder.elements[len(builder.elements)-1]
}

// SetLast replaces the last element with the provided one.
func (builder *TxDataBuilder) SetLast(element string) {
	if len(builder.elements) == 0 {
		builder.elements = []string{element}
	}

	builder.elements[len(builder.elements)-1] = element
}

// Func sets the function to be invoked by the data string.
func (builder *TxDataBuilder) Func(function string) *TxDataBuilder {
	builder.function = function

	return builder
}

// Byte appends a single byte to the data string.
func (builder *TxDataBuilder) Byte(value byte) *TxDataBuilder {
	element := hex.EncodeToString([]byte{value})
	builder.elements = append(builder.elements, element)

	return builder
}

// Bytes appends a slice of bytes to the data string.
func (builder *TxDataBuilder) Bytes(bytes []byte) *TxDataBuilder {
	element := hex.EncodeToString(bytes)
	builder.elements = append(builder.elements, element)

	return builder
}

// Str appends a string to the data string.
func (builder *TxDataBuilder) Str(str string) *TxDataBuilder {
	element := hex.EncodeToString([]byte(str))
	builder.elements = append(builder.elements, element)

	return builder
}

// Int appends an integer to the data string.
func (builder *TxDataBuilder) Int(value int) *TxDataBuilder {
	element := hex.EncodeToString(big.NewInt(int64(value)).Bytes())
	builder.elements = append(builder.elements, element)

	return builder
}

// Int64 appends an int64 to the data string.
func (builder *TxDataBuilder) Int64(value int64) *TxDataBuilder {
	element := hex.EncodeToString(big.NewInt(value).Bytes())
	builder.elements = append(builder.elements, element)

	return builder
}

// True appends the string "true" to the data string.
func (builder *TxDataBuilder) True() *TxDataBuilder {
	return builder.Str("true")
}

// False appends the string "false" to the data string.
func (builder *TxDataBuilder) False() *TxDataBuilder {
	return builder.Str("false")
}

// Bool appends either "true" or "false" to the data string, depending on the
// `value` argument.
func (builder *TxDataBuilder) Bool(value bool) *TxDataBuilder {
	if value {
		return builder.True()
	}

	return builder.False()
}

// BigInt appends the bytes of a big.Int to the data string.
func (builder *TxDataBuilder) BigInt(value *big.Int) *TxDataBuilder {
	return builder.Bytes(value.Bytes())
}

// IssueESDT appends to the data string all the elements required to request an ESDT issuing.
func (builder *TxDataBuilder) IssueESDT(token string, ticker string, supply int64, numDecimals byte) *TxDataBuilder {
	return builder.Func("issue").Str(token).Str(ticker).Int64(supply).Byte(numDecimals)
}

// TransferESDT appends to the data string all the elements required to request an ESDT transfer.
func (builder *TxDataBuilder) TransferESDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTTransfer).Str(token).Int64(value)
}

//TransferESDTNFT appends to the data string all the elements required to request an ESDT NFT transfer.
func (builder *TxDataBuilder) TransferESDTNFT(token string, nonce int, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTNFTTransfer).Str(token).Int(nonce).Int64(value)
}

// BurnESDT appends to the data string all the elements required to burn ESDT tokens.
func (builder *TxDataBuilder) BurnESDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTBurn).Str(token).Int64(value)
}

// LocalBurnESDT appends to the data string all the elements required to local burn ESDT tokens.
func (builder *TxDataBuilder) LocalBurnESDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTLocalBurn).Str(token).Int64(value)
}

// CanFreeze appends "canFreeze" followed by the provided boolean value.
func (builder *TxDataBuilder) CanFreeze(prop bool) *TxDataBuilder {
	return builder.Str("canFreeze").Bool(prop)
}

// CanWipe appends "canWipe" followed by the provided boolean value.
func (builder *TxDataBuilder) CanWipe(prop bool) *TxDataBuilder {
	return builder.Str("canWipe").Bool(prop)
}

// CanPause appends "canPause" followed by the provided boolean value.
func (builder *TxDataBuilder) CanPause(prop bool) *TxDataBuilder {
	return builder.Str("canPause").Bool(prop)
}

// CanMint appends "canMint" followed by the provided boolean value.
func (builder *TxDataBuilder) CanMint(prop bool) *TxDataBuilder {
	return builder.Str("canMint").Bool(prop)
}

// CanBurn appends "canBurn" followed by the provided boolean value.
func (builder *TxDataBuilder) CanBurn(prop bool) *TxDataBuilder {
	return builder.Str("canBurn").Bool(prop)
}

// CanTransferNFTCreateRole appends "canTransferNFTCreateRole" followed by the provided boolean value.
func (builder *TxDataBuilder) CanTransferNFTCreateRole(prop bool) *TxDataBuilder {
	return builder.Str("canTransferNFTCreateRole").Bool(prop)
}

// CanAddSpecialRoles appends "canAddSpecialRoles" followed by the provided boolean value.
func (builder *TxDataBuilder) CanAddSpecialRoles(prop bool) *TxDataBuilder {
	return builder.Str("canAddSpecialRoles").Bool(prop)
}

// IsInterfaceNil returns true if there is no value under the interface
func (builder *TxDataBuilder) IsInterfaceNil() bool {
	return builder == nil
}

func (builder *TxDataBuilder) toStringWithFunction() string {
	data := builder.function
	for _, element := range builder.elements {
		data = data + builder.separator + element
	}

	return data
}

func (builder *TxDataBuilder) toStringWithoutFunction() string {
	data := ""
	for i, element := range builder.elements {
		if i == 0 {
			data = element
			continue
		}
		data = data + builder.separator + element
	}

	return data
}
