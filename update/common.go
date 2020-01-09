package update

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/state/factory"
)

func CreateTrieIdentifier(shID uint32, accountType factory.Type) string {
	return fmt.Sprint("%d %d", shID, accountType)
}
