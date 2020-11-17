package storageResolvers

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// maxBuffToSendTrieNodes represents max buffer size to send in bytes
var maxBuffToSendTrieNodes = uint64(1 << 18) //256KB

// ArgTrieResolver is the argument structure used to create new TrieResolver instance
type ArgTrieResolver struct {
	Messenger                dataRetriever.MessageHandler
	ResponseTopicName        string
	Marshalizer              marshal.Marshalizer
	TrieDataGetter           dataRetriever.TrieDataGetter
	TrieStorageManager       data.StorageManager
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	DelayBeforeGracefulClose time.Duration
}

type trieNodeResolver struct {
	*storageResolver
	trieDataGetter     dataRetriever.TrieDataGetter
	trieStorageManager data.StorageManager
	marshalizer        marshal.Marshalizer
}

// NewTrieNodeResolver returns a new trie node resolver instance. It uses trie snapshots in order to get older data
func NewTrieNodeResolver(arg ArgTrieResolver) *trieNodeResolver {
	//TODO (JLS, this PR) add checks
	return &trieNodeResolver{
		storageResolver: &storageResolver{
			messenger:                arg.Messenger,
			responseTopicName:        arg.ResponseTopicName,
			manualEpochStartNotifier: arg.ManualEpochStartNotifier,
			chanGracefullyClose:      arg.ChanGracefullyClose,
			delayBeforeGracefulClose: arg.DelayBeforeGracefulClose,
		},
		trieStorageManager: arg.TrieStorageManager,
		trieDataGetter:     arg.TrieDataGetter,
		marshalizer:        arg.Marshalizer,
	}
}

// RequestDataFromHash tries to fetch the required trie node and send it to self
func (tnr *trieNodeResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	nodes, _, err := tnr.getSubTrie(hash, maxBuffToSendTrieNodes)
	if err != nil {
		return err
	}

	return tnr.sendDataToSelf(nodes)
}

// RequestDataFromHashArray tries to fetch the required trie nodes and send it to self
func (tnr *trieNodeResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
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

func (tnr *trieNodeResolver) getSubTrie(hash []byte, remainingSpace uint64) ([][]byte, uint64, error) {
	serializedNodes, remainingSpace, err := tnr.trieDataGetter.GetSerializedNodes(hash, remainingSpace)
	if err != nil {
		tnr.signalGracefullyClose()

		return nil, remainingSpace, err
	}

	return serializedNodes, remainingSpace, nil
}

func (tnr *trieNodeResolver) sendDataToSelf(serializedNodes [][]byte) error {
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
func (tnr *trieNodeResolver) Close() error {
	return tnr.trieStorageManager.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnr *trieNodeResolver) IsInterfaceNil() bool {
	return tnr == nil
}
