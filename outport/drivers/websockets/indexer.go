package websockets

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/websockets/client"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/websockets/process"
	"github.com/ElrondNetwork/elrond-go/outport/types"
	"github.com/gorilla/websocket"
)

// ArgsWebSocketIndexer -
type ArgsWebSocketIndexer struct {
	Hasher                   hashing.Hasher
	Marshalizer              marshal.Marshalizer
	ValidatorPubkeyConverter core.PubkeyConverter
}

type webSocketsIndexer struct {
	client       ClientHandler
	dataPrepared DataPreparer
}

// NewWebSocketIndexer -
func NewWebSocketIndexer(args *ArgsWebSocketIndexer) (*webSocketsIndexer, error) {
	argsDataWriter := &process.ArgsDataWriter{
		Hasher:                   args.Hasher,
		Marshalizer:              args.Marshalizer,
		ValidatorPubkeyConverter: args.ValidatorPubkeyConverter,
	}

	dp, err := process.NewDataWriter(argsDataWriter)
	if err != nil {
		return nil, err
	}

	cl, err := client.NewWebSocketClient(dp)
	if err != nil {
		return nil, err
	}

	return &webSocketsIndexer{
		client:       cl,
		dataPrepared: dp,
	}, nil
}

// StartSendingBlocking -
func (wsi *webSocketsIndexer) StartSendingBlocking(conn *websocket.Conn) {
	wsi.client.StartSendingBlocking(conn)
}

// SaveBlock -
func (wsi *webSocketsIndexer) SaveBlock(args types.ArgsSaveBlocks) {
	_ = wsi.dataPrepared.SaveBlock(args)
}

// SaveValidatorsPubKeys -
func (wsi *webSocketsIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	_ = wsi.dataPrepared.SaveValidatorsPubKeys(validatorsPubKeys, epoch)
}

// RevertBlock -
func (wsi *webSocketsIndexer) RevertBlock(header data.HeaderHandler, body data.BodyHandler) {
	panic("implement me")
}

// SaveRoundsInfo -
func (wsi *webSocketsIndexer) SaveRoundsInfo(roundsInfos []types.RoundInfo) {
	panic("implement me")
}

// UpdateTPS -
func (wsi *webSocketsIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	panic("implement me")
}

// SaveValidatorsRating -
func (wsi *webSocketsIndexer) SaveValidatorsRating(indexID string, infoRating []types.ValidatorRatingInfo) {
	panic("implement me")
}

// SaveAccounts -
func (wsi *webSocketsIndexer) SaveAccounts(acc []state.UserAccountHandler) {
	panic("implement me")
}

// Close -
func (wsi *webSocketsIndexer) Close() error {
	panic("implement me")
}

// IsInterfaceNil -
func (wsi *webSocketsIndexer) IsInterfaceNil() bool {
	return wsi == nil
}
