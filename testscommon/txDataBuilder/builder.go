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

func NewBuilder() *txDataBuilder {
	return &txDataBuilder{
		function:  "",
		elements:  make([]string, 0),
		separator: "@",
	}
}

func (builder *txDataBuilder) Clear() *txDataBuilder {
	builder.function = ""
	builder.elements = make([]string, 0)
	return builder
}

func (builder *txDataBuilder) ToString() string {
	data := builder.function
	for _, element := range builder.elements {
		data = data + builder.separator + element
	}

	return data
}

func (builder *txDataBuilder) ToBytes() []byte {
	return []byte(builder.ToString())
}

func (builder *txDataBuilder) Func(function string) *txDataBuilder {
	builder.function = function
	return builder
}

func (builder *txDataBuilder) Byte(value byte) *txDataBuilder {
	element := hex.EncodeToString([]byte{value})
	builder.elements = append(builder.elements, element)
	return builder
}
func (builder *txDataBuilder) Bytes(bytes []byte) *txDataBuilder {
	element := hex.EncodeToString(bytes)
	builder.elements = append(builder.elements, element)
	return builder
}

func (builder *txDataBuilder) Str(str string) *txDataBuilder {
	element := hex.EncodeToString([]byte(str))
	builder.elements = append(builder.elements, element)
	return builder
}

func (builder *txDataBuilder) Int(value int) *txDataBuilder {
	element := hex.EncodeToString(big.NewInt(int64(value)).Bytes())
	builder.elements = append(builder.elements, element)
	return builder
}

func (builder *txDataBuilder) Int64(value int64) *txDataBuilder {
	element := hex.EncodeToString(big.NewInt(value).Bytes())
	builder.elements = append(builder.elements, element)
	return builder
}

func (builder *txDataBuilder) True() *txDataBuilder {
	return builder.Str("true")
}

func (builder *txDataBuilder) False() *txDataBuilder {
	return builder.Str("false")
}

func (builder *txDataBuilder) Bool(value bool) *txDataBuilder {
	if value {
		return builder.True()
	}
	return builder.False()
}

func (builder *txDataBuilder) BigInt(value *big.Int) *txDataBuilder {
	return builder.Bytes(value.Bytes())
}

func (builder *txDataBuilder) IssueESDT(token string, ticker string, supply int64, numDecimals byte) *txDataBuilder {
	return builder.Func("issue").Str(token).Str(ticker).Int64(supply).Byte(numDecimals)
}

func (builder *txDataBuilder) TransferESDT(token string, value int64) *txDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTTransfer).Str(token).Int64(value)
}

func (builder *txDataBuilder) BurnESDT(token string, value int64) *txDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTBurn).Str(token).Int64(value)
}

func (builder *txDataBuilder) CanFreeze(prop bool) *txDataBuilder {
	return builder.Str("canFreeze").Bool(prop)
}

func (builder *txDataBuilder) CanWipe(prop bool) *txDataBuilder {
	return builder.Str("canWipe").Bool(prop)
}

func (builder *txDataBuilder) CanPause(prop bool) *txDataBuilder {
	return builder.Str("canPause").Bool(prop)
}

func (builder *txDataBuilder) CanMint(prop bool) *txDataBuilder {
	return builder.Str("canMint").Bool(prop)
}

func (builder *txDataBuilder) CanBurn(prop bool) *txDataBuilder {
	return builder.Str("canBurn").Bool(prop)
}
