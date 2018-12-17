package elrond_go_sandbox

import (
	"fmt"
	"math/big"
	"testing"
)

func Test1(t *testing.T) {
	str := "00000000000000000000000000000000"

	fmt.Printf("%x\n", []byte(str))

	aaa := []*big.Int(nil)

	aaa = append(aaa, big.NewInt(0))
}
