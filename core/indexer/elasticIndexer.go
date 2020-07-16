package indexer

import (
	"fmt"
	"io"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/elastic/go-elasticsearch/v7"
)

type elasticIndexer struct {
	*txDatabaseProcessor

	dataNodesCount uint32

	elasticClient  *elasticClient
	firehoseClient *firehoseClient
}

// NewElasticIndexer creates an elasticsearch es and handles saving
func NewElasticIndexer(arguments ElasticIndexerArgs) (ElasticIndexer, error) {
	ec, err := newElasticClient(elasticsearch.Config{
		Addresses: []string{arguments.Url},
	})
	if err != nil {
		return nil, err
	}

	fc, err := newFirehoseClient(&firehoseConfig{
		Region: awsRegion,
	})
	if err != nil {
		return nil, err
	}

	ei := &elasticIndexer{
		elasticClient: ec,
		firehoseClient: fc,
	}

	ei.txDatabaseProcessor = newTxDatabaseProcessor(
		arguments.Hasher,
		arguments.Marshalizer,
		arguments.AddressPubkeyConverter,
		arguments.ValidatorPubkeyConverter,
	)

	err = ei.init(arguments)
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

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticIndexer) SaveMiniblocks(header data.HeaderHandler, body *block.Body) {

}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticIndexer) SaveTransactions(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
) {
	txs := ei.prepareTransactionsForDatabase(body, header, txPool, selfShardID)
	buffSlice := serializeTransactions(txs, selfShardID, ei.foundedObjMap)

	for idx := range buffSlice {
		ei.firehoseClient.PutBulkRecord(&buffSlice[idx])
		//err := ei.elasticClient.DoBulkRequest(&buffSlice[idx], txIndex)
		//if err != nil {
		//	log.Warn("indexer indexing bulk of transactions",
		//		"error", err.Error())
		//	continue
		//}
	}
}

func (ei *elasticIndexer) init(arguments ElasticIndexerArgs) error {
	err := ei.setDataNodesCount()
	if err != nil {
		return err
	}

	err = ei.createIndexPolicies(arguments.IndexPolicies)
	if err != nil {
		return err
	}

	err = ei.createIndexTemplates(arguments.IndexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexes()
	if err != nil {
		return err
	}

	return nil
}

func (ei *elasticIndexer) createIndexPolicies(indexPolicies map[string]io.Reader) error {
	txp := getTemplateByName(txPolicy, indexPolicies)

	if txp != nil {
		err := ei.elasticClient.CheckAndCreatePolicy(txPolicy, txp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticIndexer) createIndexTemplates(indexTemplates map[string]io.Reader) error {
	opendistroTemplate := getTemplateByName("opendistro", indexTemplates)
	if opendistroTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate("opendistro", opendistroTemplate)
		if err != nil {
			return err
		}
	}

	txTemplate := getTemplateByName(txIndex, indexTemplates)

	if txTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate(txIndex, txTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticIndexer) createIndexes() error {
	firstTxIndexName := fmt.Sprintf("%s-000001", txIndex)
	err := ei.elasticClient.CheckAndCreateIndex(firstTxIndexName)
	if err != nil {
		return err
	}

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

func (ei *elasticIndexer) foundedObjMap(hashes []string, index string) (map[string]bool, error) {
	return nil, nil
}

func getTemplateByName(templateName string, templateList map[string]io.Reader) io.Reader {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("elasticIndexer.getTemplateByName", "could not find template", templateName)
	return nil
}
