package marshal_test

import (
	"testing"

	protobuf "github.com/ElrondNetwork/elrond-go/logger/proto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

var logLine = &protobuf.LogLineMessage{
	Message:   "test",
	LogLevel:  1,
	Args:      []string{"key", "value"},
	Timestamp: 2,
}
var marsh = marshal.ProtobufMarshalizer{}

func TestProtobufMarshalizer_Marshal(t *testing.T) {
	encNode, err := marsh.Marshal(logLine)
	assert.Nil(t, err)
	assert.NotNil(t, encNode)
}

func TestProtobufMarshalizer_MarshalWrongObj(t *testing.T) {
	obj := "elrond"
	encNode, err := marsh.Marshal(obj)
	assert.Nil(t, encNode)
	assert.Equal(t, marshal.ErrMarshallingProto, err)
}

func TestProtobufMarshalizer_Unmarshal(t *testing.T) {
	encNode, _ := marsh.Marshal(logLine)
	newLogLine := &protobuf.LogLineMessage{}

	err := marsh.Unmarshal(newLogLine, encNode)
	assert.Nil(t, err)
	assert.Equal(t, logLine.Message, newLogLine.Message)
	assert.Equal(t, logLine.LogLevel, newLogLine.LogLevel)
	assert.Equal(t, logLine.Args, newLogLine.Args)
	assert.Equal(t, logLine.Timestamp, newLogLine.Timestamp)
}

func TestProtobufMarshalizer_UnmarshalWrongObj(t *testing.T) {
	encNode, _ := marsh.Marshal(logLine)
	err := marsh.Unmarshal([]byte{}, encNode)
	assert.Equal(t, marshal.ErrUnmarshallingProto, err)
}
