package vm

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
)

// TxDataBuilder constructs a string to be used for transaction arguments
type TxDataBuilder struct {
	Function  string
	Elements  []string
	Separator string
}

func NewTxDataBuilder() *TxDataBuilder {
	return &TxDataBuilder{
		Function:  "",
		Elements:  make([]string, 0),
		Separator: "@",
	}
}

func (builder *TxDataBuilder) New() *TxDataBuilder {
	builder.Function = ""
	builder.Elements = make([]string, 0)
	return builder
}

func (builder *TxDataBuilder) ToString() string {
	data := builder.Function
	for _, element := range builder.Elements {
		data = data + builder.Separator + element
	}

	return data
}

func (builder *TxDataBuilder) ToBytes() []byte {
	return []byte(builder.ToString())
}

func (builder *TxDataBuilder) Func(function string) *TxDataBuilder {
	builder.Function = function
	return builder
}

func (builder *TxDataBuilder) Byte(value byte) *TxDataBuilder {
	element := hex.EncodeToString([]byte{value})
	builder.Elements = append(builder.Elements, element)
	return builder
}
func (builder *TxDataBuilder) Bytes(bytes []byte) *TxDataBuilder {
	element := hex.EncodeToString(bytes)
	builder.Elements = append(builder.Elements, element)
	return builder
}

func (builder *TxDataBuilder) Str(str string) *TxDataBuilder {
	element := hex.EncodeToString([]byte(str))
	builder.Elements = append(builder.Elements, element)
	return builder
}

func (builder *TxDataBuilder) Int(value int) *TxDataBuilder {
	element := hex.EncodeToString(big.NewInt(int64(value)).Bytes())
	builder.Elements = append(builder.Elements, element)
	return builder
}

func (builder *TxDataBuilder) Int64(value int64) *TxDataBuilder {
	element := hex.EncodeToString(big.NewInt(value).Bytes())
	builder.Elements = append(builder.Elements, element)
	return builder
}

func (builder *TxDataBuilder) True() *TxDataBuilder {
	return builder.Str("true")
}

func (builder *TxDataBuilder) False() *TxDataBuilder {
	return builder.Str("false")
}

func (builder *TxDataBuilder) Bool(value bool) *TxDataBuilder {
	if value {
		return builder.True()
	}
	return builder.False()
}

func (builder *TxDataBuilder) BigInt(value *big.Int) *TxDataBuilder {
	return builder.Bytes(value.Bytes())
}

func (builder *TxDataBuilder) IssueESDT(token string, ticker string, supply int64, numDecimals byte) *TxDataBuilder {
	return builder.Func("issue").Str(token).Str(ticker).Int64(supply).Byte(numDecimals)
}

func (builder *TxDataBuilder) TransferESDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTTransfer).Str(token).Int64(value)
}

func (builder *TxDataBuilder) BurnESDT(token string, value int64) *TxDataBuilder {
	return builder.Func(core.BuiltInFunctionESDTBurn).Str(token).Int64(value)
}

func (builder *TxDataBuilder) CanFreeze(prop bool) *TxDataBuilder {
	return builder.Str("canFreeze").Bool(prop)
}

func (builder *TxDataBuilder) CanWipe(prop bool) *TxDataBuilder {
	return builder.Str("canWipe").Bool(prop)
}

func (builder *TxDataBuilder) CanPause(prop bool) *TxDataBuilder {
	return builder.Str("canPause").Bool(prop)
}

func (builder *TxDataBuilder) CanMint(prop bool) *TxDataBuilder {
	return builder.Str("canMint").Bool(prop)
}

func (builder *TxDataBuilder) CanBurn(prop bool) *TxDataBuilder {
	return builder.Str("canBurn").Bool(prop)
}
