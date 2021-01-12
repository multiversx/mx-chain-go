package process

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/accounts"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type objectsMap = map[string]interface{}

//ArgElasticProcessor is struct that is used to store all components that are needed to an elastic indexer
type ArgElasticProcessor struct {
	IndexTemplates           map[string]*bytes.Buffer
	IndexPolicies            map[string]*bytes.Buffer
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	Options                  *types.Options
	DBClient                 DatabaseClientHandler
	EnabledIndexes           map[string]struct{}
	AccountsDB               state.AccountsAdapter
	Denomination             int
	TransactionFeeCalculator process.TransactionFeeCalculator
	IsInImportDBMode         bool
	ShardCoordinator         sharding.Coordinator
}

type elasticProcessor struct {
	*txDatabaseProcessor
	elasticClient  DatabaseClientHandler
	parser         *dataParser
	enabledIndexes map[string]struct{}
	accountsProc   *accounts.AccountsProcessor
}

// NewElasticProcessor creates an elasticsearch es and handles saving
func NewElasticProcessor(arguments ArgElasticProcessor) (*elasticProcessor, error) {
	err := checkArgElasticProcessor(arguments)
	if err != nil {
		return nil, err
	}

	ei := &elasticProcessor{
		elasticClient: arguments.DBClient,
		parser: &dataParser{
			hasher:      arguments.Hasher,
			marshalizer: arguments.Marshalizer,
		},
		enabledIndexes: arguments.EnabledIndexes,
		accountsProc:   accounts.NewAccountsProcessor(arguments.Denomination, arguments.Marshalizer, arguments.AddressPubkeyConverter, arguments.AccountsDB),
	}

	ei.txDatabaseProcessor = newTxDatabaseProcessor(
		arguments.Hasher,
		arguments.Marshalizer,
		arguments.AddressPubkeyConverter,
		arguments.ValidatorPubkeyConverter,
		arguments.TransactionFeeCalculator,
		arguments.IsInImportDBMode,
		arguments.ShardCoordinator,
	)

	if arguments.IsInImportDBMode {
		log.Warn("the node is in import mode! Cross shard transactions and rewards where destination shard is " +
			"not the current node's shard won't be indexed in Elastic Search")
	}

	if arguments.Options.UseKibana {
		err = ei.initWithKibana(arguments.IndexTemplates, arguments.IndexPolicies)
		if err != nil {
			return nil, err
		}
	} else {
		err = ei.initNoKibana(arguments.IndexTemplates)
		if err != nil {
			return nil, err
		}
	}

	return ei, nil
}

func checkArgElasticProcessor(arguments ArgElasticProcessor) error {
	if check.IfNil(arguments.DBClient) {
		return ErrNilDatabaseClient
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(arguments.AccountsDB) {
		return ErrNilAccountsDB
	}
	if arguments.Options == nil {
		return ErrNilOptions
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return ErrNilShardCoordinator
	}

	return nil
}

func (ei *elasticProcessor) initWithKibana(indexTemplates, _ map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	// TODO: Re-activate after we think of a solid way to handle forks+rotating indexes
	//err = ei.createIndexPolicies(indexPolicies)
	//if err != nil {
	//	return err
	//}

	err = ei.createIndexTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexes()
	if err != nil {
		return err
	}

	err = ei.createAliases()
	if err != nil {
		return err
	}

	return nil
}

func (ei *elasticProcessor) initNoKibana(indexTemplates map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexes()
	if err != nil {
		return err
	}

	err = ei.createAliases()
	if err != nil {
		return err
	}

	return nil
}

func (ei *elasticProcessor) createIndexPolicies(indexPolicies map[string]*bytes.Buffer) error {
	indexesPolicies := []string{txPolicy, blockPolicy, miniblocksPolicy, ratingPolicy, roundPolicy, validatorsPolicy,
		accountsHistoryPolicy, accountsESDTHistoryPolicy, accountsESDTIndex, receiptsPolicy, scResultsPolicy}
	for _, indexPolicyName := range indexesPolicies {
		indexPolicy := getTemplateByName(indexPolicyName, indexPolicies)
		if indexPolicy != nil {
			err := ei.elasticClient.CheckAndCreatePolicy(indexPolicyName, indexPolicy)
			if err != nil {
				log.Error("check and create policy", "policy", indexPolicy, "err", err)
				return err
			}
		}
	}

	return nil
}

func (ei *elasticProcessor) createOpenDistroTemplates(indexTemplates map[string]*bytes.Buffer) error {
	opendistroTemplate := getTemplateByName(openDistroIndex, indexTemplates)
	if opendistroTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate(openDistroIndex, opendistroTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) createIndexTemplates(indexTemplates map[string]*bytes.Buffer) error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, validatorsIndex,
		accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex, accountsESDTIndex}
	for _, index := range indexes {
		indexTemplate := getTemplateByName(index, indexTemplates)
		if indexTemplate != nil {
			err := ei.elasticClient.CheckAndCreateTemplate(index, indexTemplate)
			if err != nil {
				log.Error("check and create template", "err", err,
					"index", index)
				return err
			}
		}
	}
	return nil
}

func (ei *elasticProcessor) createIndexes() error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, validatorsIndex,
		accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex, accountsESDTIndex}
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-000001", index)
		err := ei.elasticClient.CheckAndCreateIndex(indexName)
		if err != nil {
			log.Error("check and create index", "err", err)
			return err
		}
	}
	return nil
}

func (ei *elasticProcessor) createAliases() error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex,
		validatorsIndex, accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex, accountsESDTIndex}
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-000001", index)
		err := ei.elasticClient.CheckAndCreateAlias(index, indexName)
		if err != nil {
			log.Error("check and create alias", "err", err)
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) getExistingObjMap(hashes []string, index string) (map[string]bool, error) {
	if len(hashes) == 0 {
		return make(map[string]bool), nil
	}

	response, err := ei.elasticClient.DoMultiGet(hashes, index)
	if err != nil {
		return nil, err
	}

	return getDecodedResponseMultiGet(response), nil
}

func getTemplateByName(templateName string, templateList map[string]*bytes.Buffer) *bytes.Buffer {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("elasticProcessor.getTemplateByName", "could not find template", templateName)
	return nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (ei *elasticProcessor) SaveHeader(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	txsSize int,
) error {
	if !ei.isIndexEnabled(blockIndex) {
		return nil
	}

	var buff bytes.Buffer

	serializedBlock, headerHash, err := ei.parser.getSerializedElasticBlockAndHeaderHash(header, signersIndexes, body, notarizedHeadersHashes, txsSize)
	if err != nil {
		return err
	}

	buff.Grow(len(serializedBlock))
	_, err = buff.Write(serializedBlock)
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

// RemoveHeader will remove a block from elasticsearch server
func (ei *elasticProcessor) RemoveHeader(header data.HeaderHandler) error {
	headerHash, err := core.CalculateHash(ei.marshalizer, ei.hasher, header)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoBulkRemove(blockIndex, []string{hex.EncodeToString(headerHash)})
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header data.HeaderHandler, body *block.Body) error {
	if body == nil || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil
	}

	encodedMiniblocksHashes := make([]string, 0)
	selfShardID := header.GetShardID()
	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		isDstMe := selfShardID == miniblock.ReceiverShardID
		isCrossShard := miniblock.ReceiverShardID != miniblock.SenderShardID
		if isDstMe && isCrossShard {
			continue
		}

		miniblockHash, err := core.CalculateHash(ei.marshalizer, ei.hasher, miniblock)
		if err != nil {
			log.Debug("RemoveMiniblocks cannot calculate miniblock hash",
				"error", err.Error())
			continue
		}
		encodedMiniblocksHashes = append(encodedMiniblocksHashes, hex.EncodeToString(miniblockHash))

	}

	return ei.elasticClient.DoBulkRemove(miniblocksIndex, encodedMiniblocksHashes)
}

// SetTxLogsProcessor will set tx logs processor
func (ei *elasticProcessor) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error) {
	if !ei.isIndexEnabled(miniblocksIndex) {
		return map[string]bool{}, nil
	}

	miniblocks := ei.parser.getMiniblocks(header, body)
	if len(miniblocks) == 0 {
		return make(map[string]bool), nil
	}

	buff, mbHashDb := serializeBulkMiniBlocks(header.GetShardID(), miniblocks, ei.getExistingObjMap)
	return mbHashDb, ei.elasticClient.DoBulkRequest(&buff, miniblocksIndex)
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticProcessor) SaveTransactions(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
	mbsInDb map[string]bool,
) error {
	if !ei.isIndexEnabled(txIndex) {
		return nil
	}

	txs, dbScResults, dbReceipts, alteredAccounts := ei.txDatabaseProcessor.prepareTransactionsForDatabase(body, header, txPool, selfShardID)
	buffSlice, err := serializeTransactions(txs, selfShardID, ei.getExistingObjMap, mbsInDb)
	if err != nil {
		return err
	}

	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], txIndex)
		if err != nil {
			log.Warn("indexer indexing bulk of transactions",
				"error", err.Error())
			return err
		}
	}

	err = ei.indexScResults(dbScResults)
	if err != nil {
		log.Warn("indexer indexing bulk of smart contract results",
			"error", err.Error())
	}

	err = ei.indexReceipts(dbReceipts)
	if err != nil {
		log.Warn("indexer indexing bulk of receipts",
			"error", err.Error())
	}

	return ei.indexAlteredAccounts(alteredAccounts)
}

// SaveShardStatistics will prepare and save information about a shard statistics in elasticsearch server
func (ei *elasticProcessor) SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error {
	if !ei.isIndexEnabled(tpsIndex) {
		return nil
	}

	buff := prepareGeneralInfo(tpsBenchmark)

	for _, shardInfo := range tpsBenchmark.ShardStatistics() {
		serializedShardInfo, serializedMetaInfo := serializeShardInfo(shardInfo)
		if serializedShardInfo == nil {
			continue
		}

		buff.Grow(len(serializedMetaInfo) + len(serializedShardInfo))
		_, err := buff.Write(serializedMetaInfo)
		if err != nil {
			log.Warn("elastic search: update TPS write meta", "error", err.Error())
		}
		_, err = buff.Write(serializedShardInfo)
		if err != nil {
			log.Warn("elastic search: update TPS write serialized data", "error", err.Error())
		}
	}

	return ei.elasticClient.DoBulkRequest(&buff, tpsIndex)
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []types.ValidatorRatingInfo) error {
	if !ei.isIndexEnabled(ratingIndex) {
		return nil
	}

	buffSlice, err := serializeValidatorsRating(index, validatorsRatingInfo)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], ratingIndex)
		if err != nil {
			log.Warn("indexer: indexing bulk of validators rating information",
				"index", ratingIndex,
				"error", err.Error())
			return err
		}
	}

	return nil
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *elasticProcessor) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	if !ei.isIndexEnabled(validatorsIndex) {
		return nil
	}

	var buff bytes.Buffer

	shardValPubKeys := types.ValidatorsPublicKeys{
		PublicKeys: make([]string, 0, len(shardValidatorsPubKeys)),
	}
	for _, validatorPk := range shardValidatorsPubKeys {
		strValidatorPk := ei.validatorPubkeyConverter.Encode(validatorPk)
		shardValPubKeys.PublicKeys = append(shardValPubKeys.PublicKeys, strValidatorPk)
	}

	marshalizedValidatorPubKeys, err := json.Marshal(shardValPubKeys)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators public keys")
		return err
	}

	buff.Grow(len(marshalizedValidatorPubKeys))
	_, err = buff.Write(marshalizedValidatorPubKeys)
	if err != nil {
		log.Warn("elastic search: save shard validators pub keys, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      validatorsIndex,
		DocumentID: fmt.Sprintf("%d_%d", shardID, epoch),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *elasticProcessor) SaveRoundsInfo(infos []types.RoundInfo) error {
	if !ei.isIndexEnabled(roundIndex) {
		return nil
	}

	var buff bytes.Buffer

	for _, info := range infos {
		serializedRoundInfo, meta := serializeRoundInfo(info)

		buff.Grow(len(meta) + len(serializedRoundInfo))
		_, err := buff.Write(meta)
		if err != nil {
			log.Warn("indexer: cannot write meta", "error", err.Error())
		}

		_, err = buff.Write(serializedRoundInfo)
		if err != nil {
			log.Warn("indexer: cannot write serialized round info", "error", err.Error())
		}
	}

	return ei.elasticClient.DoBulkRequest(&buff, roundIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(alteredAccounts map[string]*accounts.AlteredAccount) error {
	if !ei.isIndexEnabled(accountsIndex) {
		return nil
	}

	accountsToIndexEGLD, accountsToIndexESDT := ei.accountsProc.GetAccounts(alteredAccounts)

	err := ei.SaveAccounts(accountsToIndexEGLD)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDT(accountsToIndexESDT)
}

func (ei *elasticProcessor) saveAccountsESDT(wrappedAccounts []*accounts.AccountESDT) error {
	if !ei.isIndexEnabled(accountsESDTIndex) {
		return nil
	}

	accountsESDTMap := ei.accountsProc.PrepareAccountsMapESDT(wrappedAccounts)

	err := ei.serializeAndIndexAccounts(accountsESDTMap, accountsESDTIndex, true)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDTHistory(accountsESDTMap)
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (ei *elasticProcessor) SaveAccounts(accounts []state.UserAccountHandler) error {
	if !ei.isIndexEnabled(accountsIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsMapEGLD(accounts)
	err := ei.serializeAndIndexAccounts(accountsMap, accountsIndex, false)
	if err != nil {
		return err
	}

	return ei.saveAccountsHistory(accountsMap)
}

func (ei *elasticProcessor) serializeAndIndexAccounts(accountsMap map[string]*types.AccountInfo, index string, areESDTAccounts bool) error {
	buffSlice, err := accounts.SerializeAccounts(accountsMap, bulkSizeThreshold, areESDTAccounts)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], index)
		if err != nil {
			log.Warn("indexer: indexing bulk of accounts",
				"index", index,
				"error", err.Error())
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) saveAccountsESDTHistory(accountsInfoMap map[string]*types.AccountInfo) error {
	if !ei.isIndexEnabled(accountsESDTHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, accountsESDTHistoryIndex)
}

func (ei *elasticProcessor) saveAccountsHistory(accountsInfoMap map[string]*types.AccountInfo) error {
	if !ei.isIndexEnabled(accountsHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, accountsHistoryIndex)
}

func (ei *elasticProcessor) serializeAndIndexAccountsHistory(accountsMap map[string]*types.AccountBalanceHistory, index string) error {
	buffSlice, err := accounts.SerializeAccountsHistory(accountsMap, bulkSizeThreshold)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], index)
		if err != nil {
			log.Warn("indexer: indexing bulk of accounts history",
				"error", err.Error())
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) indexScResults(scrs []*types.ScResult) error {
	if !ei.isIndexEnabled(scResultsIndex) {
		return nil
	}

	buffSlice, err := serializeScResults(scrs)
	if err != nil {
		return err
	}

	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], scResultsIndex)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) indexReceipts(receipts []*types.Receipt) error {
	if !ei.isIndexEnabled(scResultsIndex) {
		return nil
	}

	buffSlice, err := serializeReceipts(receipts)
	if err != nil {
		return err
	}

	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], receiptsIndex)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) isIndexEnabled(index string) bool {
	_, isEnabled := ei.enabledIndexes[index]
	return isEnabled
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticProcessor) IsInterfaceNil() bool {
	return ei == nil
}
