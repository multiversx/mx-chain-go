package process

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type marshalizedReceiptsProvider interface {
	CreateMarshalizedReceipts() ([]byte, error)
}

// ArgsSaveReceipts holds arguments for creating a receiptsSaver
type ArgsSaveReceipts struct {
	MarshalizedReceiptsProvider marshalizedReceiptsProvider
	Marshaller                  marshal.Marshalizer
	Hasher                      hashing.Hasher
	Store                       dataRetriever.StorageService
	Header                      data.HeaderHandler
	HeaderHash                  []byte
}

// SaveReceipts saves receipts in the storer (receipts unit)
func SaveReceipts(args ArgsSaveReceipts) {
	marshalizedReceipts, errNotCritical := args.MarshalizedReceiptsProvider.CreateMarshalizedReceipts()
	if errNotCritical != nil {
		log.Warn("SaveReceipts(), error on CreateMarshalizedReceipts()", "error", errNotCritical.Error())
		return
	}

	if len(marshalizedReceipts) == 0 {
		return
	}

	emptyReceipts := &batch.Batch{Data: make([][]byte, 0)}
	emptyReceiptsHash, errNotCritical := core.CalculateHash(args.Marshaller, args.Hasher, emptyReceipts)
	if errNotCritical != nil {
		log.Warn("SaveReceipts(), error on CalculateHash(emptyReceipts)", "error", errNotCritical.Error())
		return
	}

	storerKey := args.Header.GetReceiptsHash()
	if bytes.Equal(storerKey, emptyReceiptsHash) {
		storerKey = args.HeaderHash
	}

	log.Debug("SaveReceipts()", "blockNonce", args.Header.GetNonce(), "storerKey", storerKey)
	errNotCritical = args.Store.Put(dataRetriever.ReceiptsUnit, storerKey, marshalizedReceipts)
	if errNotCritical != nil {
		log.Warn("SaveReceipts(), error on Store.Put()", "error", errNotCritical.Error())
	}
}
