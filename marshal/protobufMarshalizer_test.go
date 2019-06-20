package marshal_test

import (
	"testing"

	protobuf "github.com/ElrondNetwork/elrond-go/data/trie2/proto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

var node = &protobuf.CollapsedBn{EncodedChildren: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}}
var marsh = marshal.ProtobufMarshalizer{}

func TestProtobufMarshalizer_Marshal(t *testing.T) {
	encNode, err := marsh.Marshal(node)
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
	encNode, _ := marsh.Marshal(node)
	newNode := &protobuf.CollapsedBn{EncodedChildren: [][]byte{{}, {}, {}}}

	err := marsh.Unmarshal(newNode, encNode)
	assert.Nil(t, err)
	assert.Equal(t, node.EncodedChildren, newNode.EncodedChildren)
}

func TestProtobufMarshalizer_UnmarshalWrongObj(t *testing.T) {
	encNode, _ := marsh.Marshal(node)
	err := marsh.Unmarshal([]byte{}, encNode)
	assert.Equal(t, marshal.ErrUnmarshallingProto, err)
}
