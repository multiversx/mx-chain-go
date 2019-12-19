package marshal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	IntField       int    `json:"int_field"`
	BoolField      bool   `json:"bool_field"`
	StringField    string `json:"string_field"`
	BytesField     []byte `json:"bytes_field"`
	OptStringField string `json:"opt_string_field,omitempty"`
	OptBytesField  []byte `json:"opt_bytes_field,omitempty"`
}

const (
	goodFull = `{
		"int_field":10,
		"bool_field":true,
		"string_field":"some string",
		"bytes_field":"AQIDBAUG",
		"opt_string_field":"some optional string",
		"opt_bytes_field":"AQIDBAUGBwgJ"
	}`
	good = `{
		"int_field":10,
		"bool_field":true,
		"string_field":"some string",
		"bytes_field":"AQIDBAUG"
	}`

	withExtra = `{
		"int_field":10,
		"bool_field":true,
		"string_field":"some string",
		"bytes_field":"AQIDBAUG",
		"extra_string_field":"some optional string",
		"extra_string_field1":"some optional string",
		"extra_bytes_field":"AQIDBAUGBwgJ",
		"extra_bytes_field1":"AQIDBAUGBwgJ"
	}`

	badSyntax = `{
		"bool_field":true,
		"string_field":"some string",
		"bytes_field":"AQIDBAUG",
	}`
)

func TestSizeUnmarshlizer(t *testing.T) {
	jm := &JsonMarshalizer{}
	m := NewSizeCheckUnmarshalizer(jm, 20)

	ts := &testStruct{}

	err := m.Unmarshal(ts, []byte(goodFull))
	assert.Nil(t, err)

	err = m.Unmarshal(ts, []byte(good))
	assert.Nil(t, err)

	err = m.Unmarshal(ts, []byte(badSyntax))
	assert.NotNil(t, err)

	err = m.Unmarshal(ts, []byte(withExtra))
	assert.NotNil(t, err)
}

func TestSizeUnmarshlizer_BadConstruction(t *testing.T) {
	var jm *JsonMarshalizer
	m := NewSizeCheckUnmarshalizer(jm, 20)
	var scu *sizeCheckUnmarshalizer
	m2 := scu

	assert.True(t, m.IsInterfaceNil())
	assert.True(t, m2.IsInterfaceNil())
}

func TestSizeUnmarshlizer_MU(t *testing.T) {
	jm := &JsonMarshalizer{}
	m := NewSizeCheckUnmarshalizer(jm, 20)

	o := testStruct{
		IntField:       1,
		BoolField:      true,
		StringField:    "test",
		BytesField:     []byte{1, 2, 3, 4, 5, 6},
		OptStringField: "opt str 1",
		OptBytesField:  nil,
	}
	o2 := testStruct{}
	assert.NotEqual(t, o, o2)
	bytes, err := m.Marshal(&o)
	assert.Nil(t, err)
	err = m.Unmarshal(&o2, bytes)
	assert.Nil(t, err)
	assert.Equal(t, o, o2)
}

func BenchmarkSizeCheck_Disabled(b *testing.B) {
	m := &JsonMarshalizer{}
	benchInput := [][]byte{[]byte(goodFull), []byte(good), []byte(withExtra), []byte(badSyntax)}
	ts := &testStruct{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.Unmarshal(ts, benchInput[i%len(benchInput)])
	}
}

func BenchmarkSizeCheck_Enabled(b *testing.B) {
	jm := &JsonMarshalizer{}
	m := NewSizeCheckUnmarshalizer(jm, 20)
	benchInput := [][]byte{[]byte(goodFull), []byte(good), []byte(withExtra), []byte(badSyntax)}
	ts := &testStruct{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.Unmarshal(ts, benchInput[i%len(benchInput)])
	}
}
