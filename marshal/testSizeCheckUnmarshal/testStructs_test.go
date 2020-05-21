//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. testStruct.proto

package testSizeCheckUnmarshal

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshlizerProtoObj(t *testing.T) {
	t.Parallel()

	gm := &marshal.GogoProtoMarshalizer{}

	ts2 := &TestStruct2{
		Field1: 1,
		Field2: []byte("field2"),
		Field3: []byte("field3"),
	}
	ts1 := &TestStruct1{}

	marshalizedTs2, _ := gm.Marshal(ts2)
	err := gm.Unmarshal(ts1, marshalizedTs2)
	assert.Nil(t, err)
}

func TestSizeUnmarshalizerTestStruct2ToTestStruct1Err(t *testing.T) {
	t.Parallel()

	gm := &marshal.GogoProtoMarshalizer{}
	sizeCheckMarslalizer := marshal.NewSizeCheckUnmarshalizer(gm, 20)

	ts2 := &TestStruct2{
		Field1: 1,
		Field2: []byte("field2"),
		Field3: []byte("field3"),
	}
	ts1 := &TestStruct1{}

	marshalizedTs2, _ := sizeCheckMarslalizer.Marshal(ts2)
	err := sizeCheckMarslalizer.Unmarshal(ts1, marshalizedTs2)
	assert.Equal(t, marshal.ErrUnmarshallingBadSize, err)
}

func TestSizeUnmarshalizerTestStruct1ToTestStruct2(t *testing.T) {
	t.Parallel()

	gm := &marshal.GogoProtoMarshalizer{}
	sizeCheckMarslalizer := marshal.NewSizeCheckUnmarshalizer(gm, 20)

	ts1 := &TestStruct1{
		Field1: 1,
		Field2: []byte("filed2"),
	}
	ts2 := &TestStruct2{}

	marshalizedTs1, _ := sizeCheckMarslalizer.Marshal(ts1)
	err := sizeCheckMarslalizer.Unmarshal(ts2, marshalizedTs1)
	assert.Nil(t, err)
}
