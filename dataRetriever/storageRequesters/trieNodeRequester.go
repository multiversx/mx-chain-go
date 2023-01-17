package storagerequesters

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// maxBuffToSendTrieNodes represents max buffer size to send in bytes
var maxBuffToSendTrieNodes = uint64(1 << 18) //256KB

// ArgTrieRequester is the argument structure used to create new TrieRequester instance
type ArgTrieRequester struct {
	Messenger                dataRetriever.MessageHandler
	ResponseTopicName        string
	Marshalizer              marshal.Marshalizer
	TrieDataGetter           dataRetriever.TrieDataGetter
	TrieStorageManager       common.StorageManager
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	DelayBeforeGracefulClose time.Duration
}

type trieNodeRequester struct {
	*storageRequester
	trieDataGetter     dataRetriever.TrieDataGetter
	trieStorageManager common.StorageManager
	marshalizer        marshal.Marshalizer
}

// NewTrieNodeRequester returns a new trie node Requester instance. It uses trie snapshots in order to get older data
func NewTrieNodeRequester(arg ArgTrieRequester) (*trieNodeRequester, error) {
	if check.IfNil(arg.Messenger) {
		return nil, dataRetriever.ErrNilMessenger
	}
	if check.IfNil(arg.ManualEpochStartNotifier) {
		return nil, dataRetriever.ErrNilManualEpochStartNotifier
	}
	if arg.ChanGracefullyClose == nil {
		return nil, dataRetriever.ErrNilGracefullyCloseChannel
	}
	if check.IfNil(arg.TrieStorageManager) {
		return nil, dataRetriever.ErrNilTrieStorageManager
	}
	if check.IfNil(arg.TrieDataGetter) {
		return nil, dataRetriever.ErrNilTrieDataGetter
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}

	return &trieNodeRequester{
		storageRequester: &storageRequester{
			messenger:                arg.Messenger,
			responseTopicName:        arg.ResponseTopicName,
			manualEpochStartNotifier: arg.ManualEpochStartNotifier,
			chanGracefullyClose:      arg.ChanGracefullyClose,
			delayBeforeGracefulClose: arg.DelayBeforeGracefulClose,
		},
		trieStorageManager: arg.TrieStorageManager,
		trieDataGetter:     arg.TrieDataGetter,
		marshalizer:        arg.Marshalizer,
	}, nil
}

// RequestDataFromHash tries to fetch the required trie node and send it to self
func (tnr *trieNodeRequester) RequestDataFromHash(hash []byte, _ uint32) error {
	nodes, _, err := tnr.getSubTrie(hash, maxBuffToSendTrieNodes)
	if err != nil {
		return err
	}

	return tnr.sendDataToSelf(nodes)
}

// RequestDataFromHashArray tries to fetch the required trie nodes and send it to self
func (tnr *trieNodeRequester) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
	remainingSpace := maxBuffToSendTrieNodes
	nodes := make([][]byte, 0, maxBuffToSendTrieNodes)
	var nextNodes [][]byte
	var err error
	for _, hash := range hashes {
		nextNodes, remainingSpace, err = tnr.getSubTrie(hash, remainingSpace)
		if err != nil {
			continue
		}

		nodes = append(nodes, nextNodes...)

		lenNextNodes := uint64(len(nextNodes))
		if lenNextNodes == 0 || remainingSpace == 0 {
			break
		}
	}

	return tnr.sendDataToSelf(nodes)
}

func (tnr *trieNodeRequester) getSubTrie(hash []byte, remainingSpace uint64) ([][]byte, uint64, error) {
	serializedNodes, remainingSpace, err := tnr.trieDataGetter.GetSerializedNodes(hash, remainingSpace)
	if err != nil {
		tnr.signalGracefullyClose()
		return nil, remainingSpace, err
	}

	return serializedNodes, remainingSpace, nil
}

func (tnr *trieNodeRequester) sendDataToSelf(serializedNodes [][]byte) error {
	buff, err := tnr.marshalizer.Marshal(
		&batch.Batch{
			Data: serializedNodes,
		})
	if err != nil {
		return err
	}

	return tnr.sendToSelf(buff)
}

// Close will try to close the associated opened storers
func (tnr *trieNodeRequester) Close() error {
	var err error
	if !check.IfNil(tnr.trieStorageManager) {
		err = tnr.trieStorageManager.Close()
	}
	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnr *trieNodeRequester) IsInterfaceNil() bool {
	return tnr == nil
}
