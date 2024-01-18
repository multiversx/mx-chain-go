package toml

type Config struct {
	TestConfigI8
	TestConfigI16
	TestConfigI32
	TestConfigI64
	TestConfigU8
	TestConfigU16
	TestConfigU32
	TestConfigU64
	TestConfigF32
	TestConfigF64
	TestConfigStruct
	TestConfigNestedStruct
}

type TestConfigI8 struct {
	Int8 Int8
}

type Int8 struct {
	Value int8
}

type TestConfigI16 struct {
	Int16
}

type Int16 struct {
	Value int16
}

type TestConfigI32 struct {
	Int32
}

type Int32 struct {
	Value int32
}

type TestConfigI64 struct {
	Int64
}

type Int64 struct {
	Value int64
}

type TestConfigU8 struct {
	Uint8
}

type Uint8 struct {
	Value uint8
}

type TestConfigU16 struct {
	Uint16
}

type Uint16 struct {
	Value uint16
}

type TestConfigU32 struct {
	Uint32
}

type Uint32 struct {
	Value uint32
}

type TestConfigU64 struct {
	Uint64
}

type Uint64 struct {
	Value uint64
}

type TestConfigF32 struct {
	Float32
}

type Float32 struct {
	Value float32
}

type TestConfigF64 struct {
	Float64
}

type Float64 struct {
	Value float64
}

type TestConfigStruct struct {
	ConfigStruct
}

type ConfigStruct struct {
	Title string
	Description
}

type Description struct {
	Number uint32
}

type TestConfigNestedStruct struct {
	ConfigNestedStruct
}

type ConfigNestedStruct struct {
	Text string
	Message
}

type Message struct {
	Public             bool
	MessageDescription []MessageDescription
}

type MessageDescription struct {
	Text string
}
