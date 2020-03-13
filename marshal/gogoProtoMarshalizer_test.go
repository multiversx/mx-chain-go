package marshal_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

var miniblock = &block.MiniBlock{}
var gogoMarsh = marshal.GogoProtoMarshalizer{}

func recovedMarshal(obj interface{}) (buf []byte, err error) {
	defer func() {
		if p := recover(); p != nil {
			if panicError, ok := p.(error); ok {
				err = panicError
			} else {
				err = fmt.Errorf("%#v", p)
			}
			buf = nil
		}
	}()
	buf, err = gogoMarsh.Marshal(obj)
	return
}

func recovedUnmarshal(obj interface{}, buf []byte) (err error) {
	defer func() {
		if p := recover(); p != nil {
			if panicError, ok := p.(error); ok {
				err = panicError
			} else {
				err = fmt.Errorf("%#v", p)
			}
		}
	}()
	err = gogoMarsh.Unmarshal(obj, buf)
	return
}

func TestGogoProtoMarshalizer_Marshal(t *testing.T) {
	encNode, err := recovedMarshal(miniblock)
	assert.Nil(t, err)
	assert.NotNil(t, encNode)
}

func TestGogoProtoMarshalizer_MarshalWrongObj(t *testing.T) {

	obj := "elrond"
	encNode, err := recovedMarshal(obj)
	assert.Nil(t, encNode)
	assert.NotNil(t, err)
}

func TestGogoProtoMarshalizer_Unmarshal(t *testing.T) {
	encNode, _ := gogoMarsh.Marshal(miniblock)
	newNode := &block.MiniBlock{}

	err := recovedUnmarshal(newNode, encNode)
	assert.Nil(t, err)
	assert.Equal(t, miniblock, newNode)
}

func TestGogoProtoMarshalizer_UnmarshalWrongObj(t *testing.T) {
	encNode, _ := gogoMarsh.Marshal(miniblock)
	err := recovedUnmarshal([]byte{}, encNode)
	assert.NotNil(t, err)
}
