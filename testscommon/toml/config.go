package toml

// Config will hold the testing configuration parameters
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
	TestMap
	TestInterface
	TestArray
}

// TestConfigI8 will hold an int8 value for testing
type TestConfigI8 struct {
	Int8 Int8
}

// Int8 will hold the value
type Int8 struct {
	Number int8
}

// TestConfigI16 will hold an int16 value for testing
type TestConfigI16 struct {
	Int16
}

// Int16 will hold the value
type Int16 struct {
	Number int16
}

// TestConfigI32 will hold an int32 value for testing
type TestConfigI32 struct {
	Int32
}

// Int32 will hold the value
type Int32 struct {
	Number int32
}

// TestConfigI64 will hold an int64 value for testing
type TestConfigI64 struct {
	Int64
}

// Int64 will hold the value
type Int64 struct {
	Number int64
}

// TestConfigU8 will hold an uint8 value for testing
type TestConfigU8 struct {
	Uint8
}

// Uint8 will hold the value
type Uint8 struct {
	Number uint8
}

// TestConfigU16 will hold an uint16 value for testing
type TestConfigU16 struct {
	Uint16
}

// Uint16 will hold the value
type Uint16 struct {
	Number uint16
}

// TestConfigU32 will hold an uint32 value for testing
type TestConfigU32 struct {
	Uint32
}

// Uint32 will hold the value
type Uint32 struct {
	Number uint32
}

// TestConfigU64 will hold an uint64 value for testing
type TestConfigU64 struct {
	Uint64
}

// Uint64 will hold the value
type Uint64 struct {
	Number uint64
}

// TestConfigF32 will hold a float32 value for testing
type TestConfigF32 struct {
	Float32
}

// Float32 will hold the value
type Float32 struct {
	Number float32
}

// TestConfigF64 will hold a float64 value for testing
type TestConfigF64 struct {
	Float64
}

// Float64 will hold the value
type Float64 struct {
	Number float64
}

// TestConfigStruct will hold a configuration struct for testing
type TestConfigStruct struct {
	ConfigStruct
}

// ConfigStruct will hold a struct for testing
type ConfigStruct struct {
	Title string
	Description
}

// Description will hold the number
type Description struct {
	Number uint32
}

// TestConfigNestedStruct will hold a configuration with nested struct for testing
type TestConfigNestedStruct struct {
	ConfigNestedStruct
}

// ConfigNestedStruct will hold a nested struct for testing
type ConfigNestedStruct struct {
	Text string
	Message
}

// Message will hold some details
type Message struct {
	Public             bool
	MessageDescription []MessageDescription
}

// MessageDescription will hold the text
type MessageDescription struct {
	Text string
}

// MessageDescriptionOtherType will hold the text as integer
type MessageDescriptionOtherType struct {
	Text int
}

// MessageDescriptionOtherName will hold the value
type MessageDescriptionOtherName struct {
	Value string
}

// TestMap will hold a map for testing
type TestMap struct {
	Map map[string]MapValues
}

// MapValues will hold a value for map
type MapValues struct {
	Number int
}

// TestInterface will hold an interface for testing
type TestInterface struct {
	Value interface{}
}

// TestArray will hold an array of strings and integers
type TestArray struct {
	Strings []string
	Ints    []int
}
