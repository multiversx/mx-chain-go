package state

import (
	"fmt"
	"testing"
)

func TestDataNodes(t *testing.T) {
	dataNodes := make([]dataNode, 4)

	dataNodes[3] = valueNode{65, 66}

	dataNodes[0] = &fullNode{Children: [17]dataNode{dataNodes[3]}}
	dataNodes[1] = &shortNode{Key: []byte{67, 68}, Val: dataNodes[3]}
	dataNodes[2] = hashNode{69, 70}

	for _, dn := range dataNodes {
		fmt.Println(dn.String())
		fmt.Printf("Can unload: %v\n", dn.canUnload(uint16(0), uint16(0)))
		h, b := dn.cache()
		fmt.Printf("Cache: %v, %v\n\n", h, b)
	}

}
