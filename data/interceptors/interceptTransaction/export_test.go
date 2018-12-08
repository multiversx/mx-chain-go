package interceptTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (txi *transactionInterceptor) ProcessTx(tx p2p.Newer, rawData []byte) bool {
	return txi.processTx(tx, rawData)
}
