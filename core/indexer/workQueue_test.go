package indexer

import (
	"fmt"
	"testing"
)

func TestBackOff(t *testing.T) {
	wq, _ := NewWorkQueue()
	for i := 0; i < 10000; i++ {
		wq.GotBackOff()
		fmt.Println(wq.backOff.Seconds())
	}
}
