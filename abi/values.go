package abi

import "math/big"

// U8Value is a wrapper for uint8
type U8Value struct {
	Value uint8
}

// U16Value is a wrapper for uint16
type U16Value struct {
	Value uint16
}

// U32Value is a wrapper for uint32
type U32Value struct {
	Value uint32
}

// U64Value is a wrapper for uint64
type U64Value struct {
	Value uint64
}

// I8Value is a wrapper for int8
type I8Value struct {
	Value int8
}

// I16Value is a wrapper for int16
type I16Value struct {
	Value int16
}

// I32Value is a wrapper for int32
type I32Value struct {
	Value int32
}

// I64Value is a wrapper for int64
type I64Value struct {
	Value int64
}

// BigIntValue is a wrapper for a big integer
type BigIntValue struct {
	Value *big.Int
}

// AddressValue is a wrapper for an address
type AddressValue struct {
	Value []byte
}

// BytesValue is a wrapper for a byte slice
type BytesValue struct {
	Value []byte
}

// StringValue is a wrapper for a string
type StringValue struct {
	Value string
}

// BoolValue is a wrapper for a boolean
type BoolValue struct {
	Value bool
}

// OptionValue is a wrapper for an optional value
type OptionValue struct {
	Value any
}

// Field is a field in a struct, enum etc.
type Field struct {
	Name  string
	Value any
}

// StructValue is a struct (collection of fields)
type StructValue struct {
	Fields []Field
}

// EnumValue is an enum (discriminant and fields)
type EnumValue struct {
	Discriminant uint8
	Fields       []Field
}

// InputListValue is a list of values (used for encoding)
type InputListValue struct {
	Items []any
}

// OutputListValue is a list of values (used for decoding)
type OutputListValue struct {
	Items       []any
	ItemCreator func() any
}

// InputMultiValue is a multi-value (used for encoding)
type InputMultiValue struct {
	Items []any
}

// OutputMultiValue is a multi-value (used for decoding)
type OutputMultiValue struct {
	Items []any
}

// InputVariadicValues holds variadic values (used for encoding)
type InputVariadicValues struct {
	Items []any
}

// OutputVariadicValues holds variadic values (used for decoding)
type OutputVariadicValues struct {
	Items       []any
	ItemCreator func() any
}
