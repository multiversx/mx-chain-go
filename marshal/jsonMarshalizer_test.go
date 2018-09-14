package marshal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type testingJM struct {
	Str string
	Val int
	Arr []int
}

var marshalizer = &JsonMarshalizer{}

func TestMarshalUnmarshal(t *testing.T) {

	tjm := testingJM{Str: "AAA", Val: 10, Arr: []int{1, 2, 3, 5, 6, 7}}

	fmt.Println("Original:")
	fmt.Println(tjm)

	buff, err := marshalizer.Marshal(tjm)

	assert.Nil(t, err)

	fmt.Printf("Marshaled with %v:\n", marshalizer.Version())
	fmt.Println(string(buff))

	tjm2 := &testingJM{}
	err = marshalizer.Unmarshal(tjm2, buff[:])

	assert.Nil(t, err)

	fmt.Println("Unmarshaled:")
	fmt.Println(*tjm2)

	assert.Equal(t, tjm, *tjm2, "Objects not equal!")
}

func TestNullsOnMarshal(t *testing.T) {
	_, err := marshalizer.Marshal(nil)

	assert.NotNil(t, err)
}

func TestNullsOnUnmarshal(t *testing.T) {
	buff := []byte{}

	//error on nil object
	err := marshalizer.Unmarshal(nil, buff)
	assert.NotNil(t, err)

	//error on nil buffer
	tjm := &testingJM{}

	err = marshalizer.Unmarshal(tjm, nil)
	assert.NotNil(t, err)

	//error on empty buff
	err = marshalizer.Unmarshal(tjm, buff[:])
	assert.NotNil(t, err)
}

func TestIntMarshaling(t *testing.T) {
	tjm := testingJM{Val: time.Now().Nanosecond()}
	tjm2 := &testingJM{}

	buff, err := marshalizer.Marshal(tjm)

	assert.Nil(t, err)

	//fmt.Println(buff)
	//fmt.Println(string(buff))

	err = marshalizer.Unmarshal(tjm2, buff)

	assert.Equal(t, tjm, *tjm2)
}

func TestConcurrency5Secs(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	counter1 := 0
	counter2 := 0

	go func() {
		defer wg.Done()

		start := time.Now()
		for time.Now().Sub(start) < time.Second*5 {
			TestIntMarshaling(t)

			counter1++
		}
	}()

	go func() {
		defer wg.Done()

		start := time.Now()
		for time.Now().Sub(start) < time.Second*5 {
			TestIntMarshaling(t)

			counter2++
		}
	}()

	wg.Wait()

	fmt.Printf("Made %d + %d concurrency tests!\n", counter1, counter2)
}
