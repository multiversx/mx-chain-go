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

func (builder *TxDataBuilder) String() string {
	data := builder.Function
	for _, element := range builder.Elements {
		data = data + builder.Separator + element
	}

	return data
}

func (builder *TxDataBuilder) Func(function string) *TxDataBuilder {
	builder.Function = function
	return builder
}

func (builder *TxDataBuilder) TransferESDT(token string, value int64) *TxDataBuilder {
	return builder.New().Func(core.BuiltInFunctionESDTTransfer).Str(token).Int64(value)
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

func (builder *TxDataBuilder) Int64(value int64) *TxDataBuilder {
	element := hex.EncodeToString(big.NewInt(value).Bytes())
	builder.Elements = append(builder.Elements, element)
	return builder
}
