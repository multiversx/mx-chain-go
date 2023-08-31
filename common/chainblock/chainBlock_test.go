package chainblock

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestChainBlock_Add(t *testing.T) {

}

// a - b -   c -  d - e - f
//
//	b1 - c1
//	     c2 - d1
func TestChainBlock_LongestChain(t *testing.T) {
	a := &testscommon.HeaderHandlerStub{}
	first := NewChainBlock(a, "a")

	getHeader := func(prevHash string) data.HeaderHandler {
		return &testscommon.HeaderHandlerStub{
			GetPrevHashCalled: func() []byte {
				return []byte(prevHash)
			},
		}
	}

	b := getHeader("b")

	c := getHeader("b")
	d := getHeader("c")
	e := getHeader("d")
	f := getHeader("e")
	b1 := getHeader("a")
	c1 := getHeader("b1")
	c2 := getHeader("b")
	d1 := getHeader("d1")

	first.Add(b, "b")
	first.Add(b1, "b1")
	first.Add(c, "c")
	first.Add(c1, "c1")
	first.Add(c2, "c2")
	first.Add(d, "d")
	first.Add(d1, "d1")
	first.Add(e, "e")
	first.Add(f, "f")

	longestCHain, height := first.LongestChain(make([]data.HeaderHandler, 0))
	fmt.Printf("the height is %d\n", height)
	for _, block := range longestCHain {
		fmt.Printf("prev hash: %s\n", block.GetPrevHash())
	}

}
