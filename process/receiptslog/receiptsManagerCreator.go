package receiptslog

import (
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	updateSync "github.com/multiversx/mx-chain-go/update/sync"
)

// ArgsCreateReceiptsManager holds all the components needed to create a receipts manager
type ArgsCreateReceiptsManager struct {
	ReceiptDataStorer   storage.Storer
	ReceiptDataCacher   storage.Cacher
	Marshaller          marshal.Marshalizer
	Hasher              hashing.Hasher
	EnableEpochsHandler common.EnableEpochsHandler
	RequestHandler      process.RequestHandler
}

// CreateReceiptsManager will create a new instance of receipts manager
func CreateReceiptsManager(args ArgsCreateReceiptsManager) (*receiptsManager, error) {
	trieHandler, err := NewTrieInteractor(ArgsTrieInteractor{
		ReceiptDataStorer:   args.ReceiptDataStorer,
		Marshaller:          args.Marshaller,
		Hasher:              args.Hasher,
		EnableEpochsHandler: args.EnableEpochsHandler,
	})
	if err != nil {
		return nil, err
	}

	receiptsDataSyncer, err := updateSync.NewReceiptsDataSyncer(updateSync.ArgsNewReceiptsDataSyncer{
		Cache:          args.ReceiptDataCacher,
		RequestHandler: args.RequestHandler,
	})
	if err != nil {
		return nil, err
	}

	return NewReceiptsManager(ArgsReceiptsManager{
		TrieHandler:        trieHandler,
		ReceiptsDataSyncer: receiptsDataSyncer,
	})
}
