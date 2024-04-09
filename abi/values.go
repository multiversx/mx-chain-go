package abi

import "math/big"

type U8Value struct {
	Value uint8
}

type U16Value struct {
	Value uint16
}

type U32Value struct {
	Value uint32
}

type U64Value struct {
	Value uint64
}

type I8Value struct {
	Value int8
}

type I16Value struct {
	Value int16
}

type I32Value struct {
	Value int32
}

type I64Value struct {
	Value int64
}

type BigIntValue struct {
	Value *big.Int
}

type AddressValue struct {
	Value []byte
}

type BytesValue struct {
	Value []byte
}

type StringValue struct {
	Value string
}

type BoolValue struct {
	Value bool
}

type OptionValue struct {
	Value any
}

type Field struct {
	Name  string
	Value any
}

type StructValue struct {
	Fields []Field
}

type TupleValue struct {
	Fields []Field
}

type EnumValue struct {
	Discriminant uint8
	Fields       []Field
}

type InputListValue struct {
	Items []any
}

type OutputListValue struct {
	Items       []any
	ItemCreator func() any
}

type InputMultiValue struct {
	Items []any
}

type InputVariadicValues struct {
	Items []any
}

type OutputMultiValue struct {
	Items []any
}

type OutputVariadicValues struct {
	Items       []any
	ItemCreator func() any
}

type OptionalValue struct {
	Value any
	IsSet bool
}
