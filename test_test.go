package elrond_go

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	a, _ := base64.StdEncoding.DecodeString("AAAAAAAAFeU=")
	val := binary.BigEndian.Uint64(a)

	fmt.Printf("%v, %d\n", a, val)

}
