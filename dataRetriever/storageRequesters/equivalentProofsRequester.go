package storagerequesters

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"time"
)

// ArgEquivalentProofsRequester is the argument structure used to create a new equivalent proofs requester instance
type ArgEquivalentProofsRequester struct {
	Messenger                dataRetriever.MessageHandler
	ResponseTopicName        string
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	DelayBeforeGracefulClose time.Duration
	NonceConverter           typeConverters.Uint64ByteSliceConverter
	Storage                  dataRetriever.StorageService
	Marshaller               marshal.Marshalizer
	EnableEpochsHandler      core.EnableEpochsHandler
}

type equivalentProofsRequester struct {
	*storageRequester
	nonceConverter      typeConverters.Uint64ByteSliceConverter
	storage             dataRetriever.StorageService
	marshaller          marshal.Marshalizer
	enableEpochsHandler core.EnableEpochsHandler
}

// NewEquivalentProofsRequester returns a new instance of equivalent proofs requester
func NewEquivalentProofsRequester(args ArgEquivalentProofsRequester) (*equivalentProofsRequester, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &equivalentProofsRequester{
		storageRequester: &storageRequester{
			messenger:                args.Messenger,
			responseTopicName:        args.ResponseTopicName,
			manualEpochStartNotifier: args.ManualEpochStartNotifier,
			chanGracefullyClose:      args.ChanGracefullyClose,
			delayBeforeGracefulClose: args.DelayBeforeGracefulClose,
		},
		nonceConverter:      args.NonceConverter,
		storage:             args.Storage,
		marshaller:          args.Marshaller,
		enableEpochsHandler: args.EnableEpochsHandler,
	}, nil
}

func checkArgs(args ArgEquivalentProofsRequester) error {
	if check.IfNil(args.Messenger) {
		return dataRetriever.ErrNilMessenger
	}
	if check.IfNil(args.ManualEpochStartNotifier) {
		return dataRetriever.ErrNilManualEpochStartNotifier
	}
	if args.ChanGracefullyClose == nil {
		return dataRetriever.ErrNilGracefullyCloseChannel
	}
	if check.IfNil(args.NonceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(args.Storage) {
		return dataRetriever.ErrNilStore
	}
	if check.IfNil(args.Marshaller) {
		return dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return dataRetriever.ErrNilEnableEpochsHandler
	}

	return nil
}

// RequestDataFromHash requests equivalent proofs data from storage for the specified hash-shard key
func (requester *equivalentProofsRequester) RequestDataFromHash(hashShardKey []byte, epoch uint32) error {
	if !requester.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return nil
	}

	headerHash, _, err := common.GetHashAndShardFromKey(hashShardKey)
	if err != nil {
		return err
	}

	equivalentProofsStorage, err := requester.storage.GetStorer(dataRetriever.ProofsUnit)
	if err != nil {
		return err
	}

	buff, err := equivalentProofsStorage.SearchFirst(headerHash)
	if err != nil {
		return err
	}

	return requester.sendToSelf(buff)
}

// RequestDataFromNonce requests equivalent proofs data from storage for the specified nonce-shard key
func (requester *equivalentProofsRequester) RequestDataFromNonce(nonceShardKey []byte, epoch uint32) error {
	if !requester.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return nil
	}

	headerNonce, shardID, err := common.GetNonceAndShardFromKey(nonceShardKey)
	if err != nil {
		return err
	}
	storer, err := requester.getStorerForShard(shardID)
	if err != nil {
		return err
	}

	nonceKey := requester.nonceConverter.ToByteSlice(headerNonce)
	hash, err := storer.SearchFirst(nonceKey)
	if err != nil {
		return err
	}

	hashShardKey := common.GetEquivalentProofHashShardKey(hash, shardID)
	return requester.RequestDataFromHash([]byte(hashShardKey), epoch)
}

func (requester *equivalentProofsRequester) getStorerForShard(shardID uint32) (storage.Storer, error) {
	if shardID == core.MetachainShardId {
		return requester.storage.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	}

	return requester.storage.GetStorer(dataRetriever.GetHdrNonceHashDataUnit(shardID))
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *equivalentProofsRequester) IsInterfaceNil() bool {
	return requester == nil
}
