package incomingHeaders

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/sovereign"
	"github.com/stretchr/testify/require"
)

func TestSovereign_IncomingHeaderHandler(t *testing.T) {
	initialBalance := big.NewInt(1000000000000)
	nodes, _, _ := sovereign.CreateGeneralSovereignSetup(initialBalance)
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	require.Equal(t, len(nodes), 3)
}
