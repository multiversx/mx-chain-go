package txDataBuilder

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
)

// txDataBuilder constructs a string to be used for transaction arguments
type txDataBuilder struct {
	function  string
	elements  []string
	separator string
}

// NewBuilder creates a new txDataBuilder instance.
func NewBuilder() *txDataBuilder {
	return &txDataBuilder{
		function:  "",
		elements:  make([]string, 0),
		separator: "@",
	}
}

// Clear resets the internal state of the txDataBuilder, allowing a new data
// string to be built.
func (builder *txDataBuilder) Clear() *txDataBuilder {
	builder.function = ""
	builder.elements = make([]string, 0)

	return builder
}

// ToString returns the data as a string.
func (builder *txDataBuilder) ToString() string {
	data := builder.function
	for _, element := range builder.elements {
		data = data + builder.separator + element
	}

	return data
}

// ToBytes returns the data as a slice of bytes.
func (builder *txDataBuilder) ToBytes() []byte {
	return []byte(builder.ToString())
}

// Func sets the function to be invoked by the data string.
func (builder *txDataBuilder) Func(function string) *txDataBuilder {
	builder.function = function

	return builder
}

// Byte appends a single byte to the data string.
func (builder *txDataBuilder) Byte(value byte) *txDataBuilder {
	element := hex.EncodeToString([]byte{value})
	builder.elements = append(builder.elements, element)

	return builder
}

// Bytes appends a slice of bytes to the data string.
func (builder *txDataBuilder) Bytes(bytes []byte) *txDataBuilder {
	element := hex.EncodeToString(bytes)
	builder.elements = append(builder.elements, element)

	return builder
}

// Str appends a string to the data string.
func (builder *txDataBuilder) Str(str string) *txDataBuilder {
	element := hex.EncodeToString([]byte(str))
	builder.elements = append(builder.elements, element)

	return builder
}

// Int appends an integer to the data string.
func (builder *txDataBuilder) Int(value int) *txDataBuilder {
	element := hex.EncodeToString(big.NewInt(int64(value)).Bytes())
	builder.elements = append(builder.elements, element)

	return builder
}

// Int64 appends an int64 to the data string.
func (builder *txDataBuilder) Int64(value int64) *txDataBuilder {
	element := hex.EncodeToString(big.NewInt(value).Bytes())
	builder.elements = append(builder.elements, element)

	return builder
}

// True appends the string "true" to the data string.
func (builder *txDataBuilder) True() *txDataBuilder {
	return builder.Str("true")
}

// False appends the string "false" to the data string.
func (builder *txDataBuilder) False() *txDataBuilder {
	return builder.Str("false")
}

// Bool appends either "true" or "false" to the data string, depending on the
// `value` argument.
func (builder *txDataBuilder) Bool(value bool) *txDataBuilder {
	if value {
		return builder.True()
	}

	return builder.False()
}

// BigInt appends the bytes of a big.Int to the data string.
func (builder *txDataBuilder) BigInt(value *big.Int) *txDataBuilder {
	return builder.Bytes(value.Bytes())
}

// IssueESDT appends to the data string all the elements required to request an ESDT issuing.
func (builder *txDataBuilder) IssueESDT(token string, ticker string, supply int64, numDecimals byte) *txDataBuilder {
	return builder.Func("issue").Str(token).Str(ticker).Int64(supply).Byte(numDecimals)
}

// TransferESDT appends to the data string all the elements required to request an ESDT transfer.
func (builder *txDataBuilder) TransferESDT(token string, value int64) *txDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTTransfer).Str(token).Int64(value)
}

// BurnESDT appends to the data string all the elements required to burn ESDT tokens.
func (builder *txDataBuilder) BurnESDT(token string, value int64) *txDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTBurn).Str(token).Int64(value)
}

// CanFreeze appends "canFreeze" followed by the provided boolean value.
func (builder *txDataBuilder) CanFreeze(prop bool) *txDataBuilder {
	return builder.Str("canFreeze").Bool(prop)
}

// CanWipe appends "canWipe" followed by the provided boolean value.
func (builder *txDataBuilder) CanWipe(prop bool) *txDataBuilder {
	return builder.Str("canWipe").Bool(prop)
}

// CanPause appends "canPause" followed by the provided boolean value.
func (builder *txDataBuilder) CanPause(prop bool) *txDataBuilder {
	return builder.Str("canPause").Bool(prop)
}

// CanMint appends "canMint" followed by the provided boolean value.
func (builder *txDataBuilder) CanMint(prop bool) *txDataBuilder {
	return builder.Str("canMint").Bool(prop)
}

// CanBurn appends "canBurn" followed by the provided boolean value.
func (builder *txDataBuilder) CanBurn(prop bool) *txDataBuilder {
	return builder.Str("canBurn").Bool(prop)
}

// IsInterfaceNil returns true if there is no value under the interface
func (builder *txDataBuilder) IsInterfaceNil() bool {
	return builder == nil
}
