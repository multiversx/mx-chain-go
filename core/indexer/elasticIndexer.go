package indexer

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/elastic/go-elasticsearch/v7"
)

type elasticIndexer struct {
	dataNodesCount uint32

	elasticClient *elasticClient
}

// NewElasticIndexer creates an elasticsearch es and handles saving
func NewElasticIndexer(arguments ElasticIndexerArgs) (ElasticIndexer, error) {
	ec, err := newElasticClient(elasticsearch.Config{
		Addresses: []string{arguments.Url},
	})
	if err != nil {
		return nil, err
	}

	ei := &elasticIndexer{
		elasticClient: ec,
	}

	err = ei.init()
	if err != nil {
		return nil, err
	}

	return ei, nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (ei *elasticIndexer) SaveHeader(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	txsSize int,
) {

}

func (ei *elasticIndexer) init() error {
	err := ei.setDataNodesCount()
	if err != nil {
		return err
	}

	err = ei.createIndexes()
	if err != nil {
		return err
	}
	return nil
}

func (ei *elasticIndexer) createIndexes() error {

	return nil
}

func (ei *elasticIndexer) setDataNodesCount() error {
	infoRes, err := ei.elasticClient.NodesInfo()
	if err != nil {
		return err
	}

	decRes := &elasticNodesInfo{}
	err = ei.elasticClient.parseResponse(infoRes, decRes)
	if err != nil {
		return err
	}

	ei.setDataNodesCountFromESResponse(decRes)

	return nil
}

func (ei *elasticIndexer) setDataNodesCountFromESResponse(eni *elasticNodesInfo) {
	for _, nodeInfo := range eni.Nodes {
		for _, nodeType := range nodeInfo.Roles {
			if nodeType == dataNodeIdentifier {
				ei.dataNodesCount++
				break
			}
		}
	}
}


