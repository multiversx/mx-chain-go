package websockets

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"net/http"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/outport/types"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestBlaBLaBLaBLa(t *testing.T) {
	t.Parallel()

	ws := gin.Default()
	ws.Use(cors.Default())
	go func() {
		_ = ws.Run(":8041")
	}()

	upgrader := websocket.Upgrader{}

	indexer, err := NewWebSocketIndexer(&ArgsWebSocketIndexer{
		Hasher:                   testscommon.HasherMock{},
		Marshalizer:              testscommon.MarshalizerMock{},
		ValidatorPubkeyConverter: testscommon.NewPubkeyConverterMock(32),
	})
	require.NoError(t, err)

	ws.GET("/indexer", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		indexer.StartSendingBlocking(conn)
	})

	args := types.ArgsSaveBlocks{
		Body: &block.Body{},
		Header: &block.Header{
			Nonce: 1,
			Round: 2,
			Epoch: 3,
		},
		TxsFromPool: map[string]data.TransactionHandler{
			"hash": &transaction.Transaction{},
		},
	}

	validatorsKey := map[uint32][][]byte{
		uint32(0): {[]byte("key10"), []byte("key20")},
		uint32(1): {[]byte("key11"), []byte("key22")},
	}

	indexer.SaveBlock(args)
	time.Sleep(time.Second)
	indexer.SaveBlock(args)
	time.Sleep(time.Second)
	indexer.SaveBlock(args)
	time.Sleep(time.Second)
	indexer.SaveBlock(args)
	time.Sleep(time.Second)
	indexer.SaveValidatorsPubKeys(validatorsKey, 0)
	time.Sleep(time.Second)

	time.Sleep(5 * time.Second)
}
