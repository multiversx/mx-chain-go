package statistics

import (
	"fmt"
	"testing"
)

func TestNewTrieStatistics(t *testing.T) {
	ts := NewTrieStatistics()
	ts.AddBranchNode(0, 15)
	println(ts)
}

func TestNewTrieStatistics2(t *testing.T) {
	a := make([]int, 10)
	for i := 9; i >= 0; i-- {
		a[9-i] = i * 2
	}

	fmt.Println(a)
	insert := 20
	lastPos := 10
	for i := 9; i >= 0; i-- {
		if a[i] < insert {
			lastPos = i
			continue
		}

	}
	if lastPos < 10 {
		//fmt.Println(a[:lastPos])
		//fmt.Println(a[lastPos : 10-1])
		//b := append(a[:lastPos], insert)
		//fmt.Println(b)
		//fmt.Println(a[lastPos : 10-1])

		a = append(a[:lastPos+1], a[lastPos:10-1]...)
		a[lastPos] = insert
	}
	fmt.Println(a)
}
