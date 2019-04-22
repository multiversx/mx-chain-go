package statistics

import (
	"fmt"
	"sync"
	"testing"
)

type X struct {
	myMap map[int]struct{}
	myMapLock sync.RWMutex
}

func TestShit(t *testing.T) {
	myX := X{
		myMap: make(map[int]struct{}),
	}
	for i := 0; i < 3; i++ {
		myX.myMap[i] = struct{}{}
	}
	delete(myX.myMap, 0)
	fmt.Println(myX.myMap)
	delete(myX.myMap, 1)
	fmt.Println(myX.myMap)
	delete(myX.myMap, 2)
	fmt.Println(myX.myMap)
}

func (x *X) OperateOnShit(i int) {
	/*str, ok := x.myMap[i]
	if !ok {
		x.myMap[i] = struct{}{}
	}*/
}