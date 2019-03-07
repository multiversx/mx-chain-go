package patriciaMerkleTrie

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func TestNewTrie(t *testing.T) {
	tr, _ := NewTrie(keccak.Keccak{}, marshal.JsonMarshalizer{}, nil)

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	proof, _ := tr.Prove([]byte("doe"))

	fmt.Println(tr.VerifyProof(proof, []byte("dog")))
}
