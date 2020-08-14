package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type elasticIndexer struct {
	*txDatabaseProcessor

	dataNodesCount uint32
	elasticClient  *elasticClient
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
) error {
	var buff bytes.Buffer

	serializedBlock, headerHash := ei.getSerializedElasticBlockAndHeaderHash(header, signersIndexes, body, notarizedHeadersHashes, txsSize)

	buff.Grow(len(serializedBlock))
	_, err := buff.Write(serializedBlock)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      blockIndex,
		DocumentID: hex.EncodeToString(headerHash),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticIndexer) SaveMiniblocks(header data.HeaderHandler, body *block.Body) map[string]bool {
	miniblocks := ei.getMiniblocks(header, body)
	if miniblocks == nil {
		log.Warn("indexer: could not index miniblocks")
		return make(map[string]bool)
	}
	if len(miniblocks) == 0 {
		return make(map[string]bool)
	}

	buff, mbHashDb := serializeBulkMiniBlocks(header.GetShardID(), miniblocks, ei.foundedObjMap)
	err := ei.elasticClient.DoBulkRequest(&buff, miniblocksIndex)
	if err != nil {
		log.Warn("indexing bulk of miniblocks", "error", err.Error())
	}

	return mbHashDb
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticIndexer) SaveTransactions(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
    mbsInDb map[string]bool,
) error {
	txs := ei.prepareTransactionsForDatabase(body, header, txPool, selfShardID)
	for _, tx := range txs {
		fmt.Println(tx.Hash)
	}
	buffSlice := serializeTransactions(txs, selfShardID, ei.foundedObjMap, mbsInDb)

	for idx := range buffSlice {
		err := ei.elasticClient.DoBulkRequest(&buffSlice[idx], txIndex)
		if err != nil {
			log.Warn("indexer indexing bulk of transactions",
				"error", err.Error())
			return err
		}
	}

	return nil
}

func (ei *elasticIndexer) init(arguments ElasticIndexerArgs) error {
	err := ei.setDataNodesCount()
	if err != nil {
		return nil
	}

	err = ei.createOpenDistroTemplates(arguments.IndexTemplates)
	if err != nil {
		return nil
	}

	err = ei.createIndexPolicies(arguments.IndexPolicies)
	if err != nil {
		return nil
	}

	err = ei.createIndexTemplates(arguments.IndexTemplates)
	if err != nil {
		return nil
	}

	err = ei.createIndexes()
	if err != nil {
		return nil
	}

	err = ei.setInitialAliases()
	if err != nil {
		return nil
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

	blockp := getTemplateByName(blockPolicy, indexPolicies)
	if blockp != nil {
		err := ei.elasticClient.CheckAndCreatePolicy(blockPolicy, blockp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticIndexer) createOpenDistroTemplates(indexTemplates map[string]io.Reader) error {
	opendistroTemplate := getTemplateByName("opendistro", indexTemplates)
	if opendistroTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate("opendistro", opendistroTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticIndexer) createIndexTemplates(indexTemplates map[string]io.Reader) error {
	txTemplate := getTemplateByName(txIndex, indexTemplates)
	if txTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate(txIndex, txTemplate)
		if err != nil {
			return err
		}
	}

	blocksTemplate := getTemplateByName(blockIndex, indexTemplates)
	if blocksTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate(blockIndex, blocksTemplate)
		if err != nil {
			return err
		}
	}

	miniblocksTemplate := getTemplateByName(miniblocksIndex, indexTemplates)
	if miniblocksTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate(miniblocksIndex, miniblocksTemplate)
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

	firstBlocksIndexName :=  fmt.Sprintf("%s-000001", blockIndex)
	err = ei.elasticClient.CheckAndCreateIndex(firstBlocksIndexName)
	if err != nil {
		return err
	}

	firstMiniBlocksIndexName :=  fmt.Sprintf("%s-000001", miniblocksIndex)
	err = ei.elasticClient.CheckAndCreateIndex(firstMiniBlocksIndexName)
	if err != nil {
		return err
	}

	return nil
}

func (ei *elasticIndexer) setInitialAliases() error {
	firstTxIndexName := fmt.Sprintf("%s-000001", txIndex)
	err := ei.elasticClient.CheckAndCreateAlias(txIndex, firstTxIndexName)
	if err != nil {
		return err
	}

	firstBlocksIndexName :=  fmt.Sprintf("%s-000001", blockIndex)
	err = ei.elasticClient.CheckAndCreateAlias(blockIndex, firstBlocksIndexName)
	if err != nil {
		return err
	}

	firstMiniBlocksIndexName :=  fmt.Sprintf("%s-000001", miniblocksIndex)
	err = ei.elasticClient.CheckAndCreateAlias(miniblocksIndex, firstMiniBlocksIndexName)
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
	err = ei.elasticClient.parseResponse(infoRes, decRes, ei.elasticClient.elasticDefaultErrorResponseHandler)
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
	if len(hashes) == 0 {
		return make(map[string]bool), nil
	}

	response, err := ei.elasticClient.DoMultiGet(getDocumentsByIDsQuery(hashes), index)
	if err != nil {
		return nil, err
	}

	return getDecodedResponseMultiGet(response), nil
}

func (ei *elasticIndexer) getMiniblocks(header data.HeaderHandler, body *block.Body) []*Miniblock {
	headerHash, err := core.CalculateHash(ei.marshalizer, ei.hasher, header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err.Error())
		return nil
	}

	encodedHeaderHash := hex.EncodeToString(headerHash)

	miniblocks := make([]*Miniblock, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(ei.marshalizer, ei.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)

		mb := &Miniblock{
			Hash:            encodedMbHash,
			SenderShardID:   miniblock.SenderShardID,
			ReceiverShardID: miniblock.ReceiverShardID,
			Type:            miniblock.Type.String(),
		}

		if mb.SenderShardID == header.GetShardID() {
			mb.SenderBlockHash = encodedHeaderHash
		} else {
			mb.ReceiverBlockHash = encodedHeaderHash
		}

		if mb.SenderShardID == mb.ReceiverShardID {
			mb.ReceiverBlockHash = encodedHeaderHash
		}

		miniblocks = append(miniblocks, mb)
	}

	return miniblocks
}

func (ei *elasticIndexer) getSerializedElasticBlockAndHeaderHash(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	sizeTxs int,
) ([]byte, []byte) {
	headerBytes, err := ei.marshalizer.Marshal(header)
	if err != nil {
		log.Debug("indexer: marshal header", "error", err)
		return nil, nil
	}
	bodyBytes, err := ei.marshalizer.Marshal(body)
	if err != nil {
		log.Debug("indexer: marshal body", "error", err)
		return nil, nil
	}

	blockSizeInBytes := len(headerBytes) + len(bodyBytes)

	miniblocksHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(ei.marshalizer, ei.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)
		miniblocksHashes = append(miniblocksHashes, encodedMbHash)
	}

	headerHash := ei.hasher.Compute(string(headerBytes))
	elasticBlock := Block{
		Nonce:                 header.GetNonce(),
		Round:                 header.GetRound(),
		Epoch:                 header.GetEpoch(),
		ShardID:               header.GetShardID(),
		Hash:                  hex.EncodeToString(headerHash),
		MiniBlocksHashes:      miniblocksHashes,
		NotarizedBlocksHashes: notarizedHeadersHashes,
		Proposer:              signersIndexes[0],
		Validators:            signersIndexes,
		PubKeyBitmap:          hex.EncodeToString(header.GetPubKeysBitmap()),
		Size:                  int64(blockSizeInBytes),
		SizeTxs:               int64(sizeTxs),
		Timestamp:             time.Duration(header.GetTimeStamp()),
		TxCount:               header.GetTxCount(),
		StateRootHash:         hex.EncodeToString(header.GetRootHash()),
		PrevHash:              hex.EncodeToString(header.GetPrevHash()),
	}

	serializedBlock, err := json.Marshal(elasticBlock)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal elastic header")
		return nil, nil
	}

	return serializedBlock, headerHash
}

func getTemplateByName(templateName string, templateList map[string]io.Reader) io.Reader {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("elasticIndexer.getTemplateByName", "could not find template", templateName)
	return nil
}
