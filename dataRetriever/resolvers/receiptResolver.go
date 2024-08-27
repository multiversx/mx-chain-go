package resolvers

import (
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgReceiptResolver is the argument structure used to create a new receiptResolver instance
type ArgReceiptResolver struct {
	ArgBaseResolver
	ReceiptPool       storage.Cacher
	ReceiptStorage    storage.Storer
	DataPacker        dataRetriever.DataPacker
	IsFullHistoryNode bool
}

// miniblockResolver is a wrapper over Resolver that is specialized in resolving receipts requests
type receiptResolver struct {
	*baseResolver
	messageProcessor
	baseStorageResolver
	receiptPool storage.Cacher
	dataPacker  dataRetriever.DataPacker
}

// NewReceiptResolver creates a receipt resolver
func NewReceiptResolver(arg ArgReceiptResolver) (*receiptResolver, error) {
	err := checkArgReceiptResolver(arg)
	if err != nil {
		return nil, err
	}

	rcResolver := &receiptResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: arg.SenderResolver,
		},
		receiptPool:         arg.ReceiptPool,
		baseStorageResolver: createBaseStorageResolver(arg.ReceiptStorage, arg.IsFullHistoryNode),
		dataPacker:          arg.DataPacker,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshaller,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.RequestTopic(),
			throttler:        arg.Throttler,
		},
	}

	return rcResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (rcRes *receiptResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error {
	err := rcRes.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	rcRes.throttler.StartProcessing()
	defer rcRes.throttler.EndProcessing()

	rd, err := rcRes.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.HashType:
		err = rcRes.resolveReceiptRequestByHash(rd.Value, message.Peer(), rd.Epoch, source)
	case dataRetriever.HashArrayType:
		err = rcRes.resolveReceiptRequestByHashArray(rd.Value, message.Peer(), rd.Epoch, source)
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}

	if err != nil {
		err = fmt.Errorf("%w for hash %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

func (rcRes *receiptResolver) resolveReceiptRequestByHash(hash []byte, pid core.PeerID, epoch uint32, source p2p.MessageHandler) error {
	receipt, err := rcRes.fetchReceiptAsByteSlice(hash, epoch)
	if err != nil {
		return err
	}

	b := &batch.Batch{
		Data: [][]byte{receipt},
	}
	buffToSend, err := rcRes.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return rcRes.Send(buffToSend, pid, source)
}

func (rcRes *receiptResolver) fetchReceiptAsByteSlice(hash []byte, epoch uint32) ([]byte, error) {
	value, ok := rcRes.receiptPool.Peek(hash)
	if ok {
		return rcRes.marshalizer.Marshal(value)
	}

	buff, err := rcRes.getFromStorage(hash, epoch)
	if err != nil {
		rcRes.DebugHandler().LogFailedToResolveData(
			rcRes.topic,
			hash,
			err,
		)

		return nil, err
	}

	rcRes.DebugHandler().LogSucceededToResolveData(rcRes.topic, hash)

	return buff, nil
}

func (rcRes *receiptResolver) resolveReceiptRequestByHashArray(mbBuff []byte, pid core.PeerID, epoch uint32, source p2p.MessageHandler) error {
	b := batch.Batch{}
	err := rcRes.marshalizer.Unmarshal(&b, mbBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	var errFetch error
	errorsFound := 0
	mbsBuffSlice := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		mb, errTemp := rcRes.fetchReceiptAsByteSlice(hash, epoch)
		if errTemp != nil {
			errFetch = fmt.Errorf("%w for hash %s", errTemp, logger.DisplayByteSlice(hash))
			log.Trace("fetchMbAsByteSlice missing",
				"error", errFetch.Error(),
				"hash", hash)
			errorsFound++

			continue
		}
		mbsBuffSlice = append(mbsBuffSlice, mb)
	}

	buffsToSend, errPack := rcRes.dataPacker.PackDataInChunks(mbsBuffSlice, maxBuffToSendBulkMiniblocks)
	if errPack != nil {
		return errPack
	}

	for _, buff := range buffsToSend {
		errSend := rcRes.Send(buff, pid, source)
		if errSend != nil {
			return errSend
		}
	}

	if errFetch != nil {
		errFetch = fmt.Errorf("resolveMbRequestByHashArray last error %w from %d encountered errors", errFetch, errorsFound)
	}

	return errFetch
}

func checkArgReceiptResolver(arg ArgReceiptResolver) error {
	err := checkArgBase(arg.ArgBaseResolver)
	if err != nil {
		return err
	}
	if check.IfNil(arg.ReceiptPool) {
		return dataRetriever.ErrNilReceiptsPool
	}
	if check.IfNil(arg.ReceiptStorage) {
		return dataRetriever.ErrNilMiniblocksStorage
	}
	if check.IfNil(arg.DataPacker) {
		return dataRetriever.ErrNilDataPacker
	}
	return nil
}
