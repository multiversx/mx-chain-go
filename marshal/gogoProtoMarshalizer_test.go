package marshal_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

var miniblock = &block.MiniBlock{}
var gogoMarsh = marshal.GogoProtoMarshalizer{}

func TestGogoProtoMarshalizer_Marshal(t *testing.T) {
	encNode, err := gogoMarsh.Marshal(miniblock)
	assert.Nil(t, err)
	assert.NotNil(t, encNode)
}

func TestGogoProtoMarshalizer_MarshalWrongObj(t *testing.T) {
	obj := "elrond"
	encNode, err := gogoMarsh.Marshal(obj)
	assert.Nil(t, encNode)
	assert.Equal(t, marshal.ErrMarshallingProto, err)
}

func TestGogoProtoMarshalizer_Unmarshal(t *testing.T) {
	encNode, _ := gogoMarsh.Marshal(miniblock)
	newNode := &block.MiniBlock{}

	err := gogoMarsh.Unmarshal(newNode, encNode)
	assert.Nil(t, err)
	assert.Equal(t, miniblock, newNode)
}

func TestGogoProtoMarshalizer_UnmarshalWrongObj(t *testing.T) {
	encNode, _ := gogoMarsh.Marshal(miniblock)
	err := gogoMarsh.Unmarshal([]byte{}, encNode)
	assert.Equal(t, marshal.ErrUnmarshallingProto, err)
}
