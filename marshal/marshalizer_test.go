package marshal_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/stretchr/testify/assert"
)

type testingJM struct {
	Str string
	Val int
	Arr []int
}

func TestJsonMarshalizer(t *testing.T) {
	Suite(t, &marshal.JsonMarshalizer{})
}

func TestDefaultMarshalizerIsNotNil(t *testing.T) {
	assert.NotNil(t, marshal.DefMarsh)
}

func Suite(t *testing.T, marshalizer marshal.Marshalizer) {
	TestingMarshalUnmarshal(t, marshalizer)
	TestingNullsOnMarshal(t, marshalizer)
	TestingNullsOnUnmarshal(t, marshalizer)
	TestingIntMarshaling(t, marshalizer)
	TestingConcurrency5Secs(t, marshalizer)
}

func TestingMarshalUnmarshal(t *testing.T, marshalizer marshal.Marshalizer) {

	tjm := testingJM{Str: "AAA", Val: 10, Arr: []int{1, 2, 3, 5, 6, 7}}

	fmt.Println("Original:")
	fmt.Println(tjm)

	buff, err := marshalizer.Marshal(tjm)

	assert.Nil(t, err)

	fmt.Println(string(buff))

	tjm2 := &testingJM{}
	err = marshalizer.Unmarshal(tjm2, buff[:])

	assert.Nil(t, err)

	fmt.Println("Unmarshaled:")
	fmt.Println(*tjm2)

	assert.Equal(t, tjm, *tjm2, "Objects not equal!")
}

func TestingNullsOnMarshal(t *testing.T, marshalizer marshal.Marshalizer) {
	_, err := marshalizer.Marshal(nil)

	assert.NotNil(t, err)
}

func TestingNullsOnUnmarshal(t *testing.T, marshalizer marshal.Marshalizer) {
	buff := make([]byte, 0)

	//error on nil object
	err := marshalizer.Unmarshal(nil, buff)
	assert.NotNil(t, err)

	//error on nil buffer
	tjm := &testingJM{}

	err = marshalizer.Unmarshal(tjm, nil)
	assert.NotNil(t, err)

	//error on empty buff
	err = marshalizer.Unmarshal(tjm, buff)
	assert.NotNil(t, err)
}

func TestingIntMarshaling(t *testing.T, marshalizer marshal.Marshalizer) {
	tjm := testingJM{Val: time.Now().Nanosecond()}
	tjm2 := &testingJM{}

	buff, err := marshalizer.Marshal(tjm)

	assert.Nil(t, err)

	err = marshalizer.Unmarshal(tjm2, buff)

	assert.Equal(t, tjm, *tjm2)
}

func TestingConcurrency5Secs(t *testing.T, marshalizer marshal.Marshalizer) {
	var wg sync.WaitGroup
	wg.Add(2)

	counter1 := 0
	counter2 := 0

	go func() {
		defer wg.Done()

		start := time.Now()
		for time.Now().Sub(start) < time.Second*5 {
			TestingIntMarshaling(t, marshalizer)

			counter1++
		}
	}()

	go func() {
		defer wg.Done()

		start := time.Now()
		for time.Now().Sub(start) < time.Second*5 {
			TestingIntMarshaling(t, marshalizer)

			counter2++
		}
	}()

	wg.Wait()

	fmt.Printf("Made %d + %d concurrency tests!\n", counter1, counter2)
}
