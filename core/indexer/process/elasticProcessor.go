package process

import (
	"bytes"
	"encoding/hex"
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

var log = logger.GetOrCreate("indexer/process")

type objectsMap = map[string]interface{}

// ArgElasticProcessor -
type ArgElasticProcessor struct {
	UseKibana       bool
	SelfShardID     uint32
	IndexTemplates  map[string]*bytes.Buffer
	IndexPolicies   map[string]*bytes.Buffer
	EnabledIndexes  map[string]struct{}
	TxProc          DBTransactionsHandler
	AccountsProc    DBAccountHandler
	BlockProc       DBBlockHandler
	MiniblocksProc  DBMiniblocksHandler
	GeneralInfoProc DBGeneralInfoHandler
	ValidatorsProc  DBValidatorsHandler
	DBClient        DatabaseClientHandler
}

type elasticProcessor struct {
	selfShardID     uint32
	enabledIndexes  map[string]struct{}
	elasticClient   DatabaseClientHandler
	accountsProc    DBAccountHandler
	blockProc       DBBlockHandler
	txProc          DBTransactionsHandler
	miniblocksProc  DBMiniblocksHandler
	generalInfoProc DBGeneralInfoHandler
	validatorsProc  DBValidatorsHandler
}

// NewElasticProcessor creates an elasticsearch es and handles saving
func NewElasticProcessor(arguments *ArgElasticProcessor) (*elasticProcessor, error) {
	ei := &elasticProcessor{
		elasticClient:   arguments.DBClient,
		enabledIndexes:  arguments.EnabledIndexes,
		accountsProc:    arguments.AccountsProc,
		blockProc:       arguments.BlockProc,
		miniblocksProc:  arguments.MiniblocksProc,
		txProc:          arguments.TxProc,
		selfShardID:     arguments.SelfShardID,
		generalInfoProc: arguments.GeneralInfoProc,
		validatorsProc:  arguments.ValidatorsProc,
	}

	if arguments.UseKibana {
		err := ei.initWithKibana(arguments.IndexTemplates, arguments.IndexPolicies)
		if err != nil {
			return nil, err
		}
	} else {
		err := ei.initNoKibana(arguments.IndexTemplates)
		if err != nil {
			return nil, err
		}
	}

	return ei, nil
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

//nolint
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
		accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex, accountsESDTIndex,
		epochInfoIndex,
	}
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
		accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex, accountsESDTIndex,
		epochInfoIndex,
	}
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
		validatorsIndex, accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex,
		accountsESDTIndex, epochInfoIndex,
	}
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
		return make(map[string]bool), err
	}

	return getDecodedResponseMultiGet(response), nil
}

func getDecodedResponseMultiGet(response objectsMap) map[string]bool {
	founded := make(map[string]bool)
	interfaceSlice, ok := response["docs"].([]interface{})
	if !ok {
		return founded
	}

	for _, element := range interfaceSlice {
		obj := element.(objectsMap)
		_, ok = obj["error"]
		if ok {
			continue
		}
		founded[obj["_id"].(string)] = obj["found"].(bool)
	}

	return founded
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

	elasticBlock, err := ei.blockProc.PrepareBlockForDB(header, signersIndexes, body, notarizedHeadersHashes, txsSize)
	if err != nil {
		return err
	}

	buff, err := ei.blockProc.SerializeBlock(elasticBlock)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      blockIndex,
		DocumentID: elasticBlock.Hash,
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	err = ei.elasticClient.DoRequest(req)
	if err != nil {
		return err
	}

	return ei.indexEpochInfoData(header)
}

func (ei *elasticProcessor) indexEpochInfoData(header data.HeaderHandler) error {
	if !ei.isIndexEnabled(epochInfoIndex) ||
		ei.selfShardID != core.MetachainShardId {
		return nil
	}

	buff, err := ei.blockProc.SerializeEpochInfoData(header)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      epochInfoIndex,
		DocumentID: fmt.Sprintf("%d", header.GetEpoch()),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// RemoveHeader will remove a block from elasticsearch server
func (ei *elasticProcessor) RemoveHeader(header data.HeaderHandler) error {
	headerHash, err := ei.blockProc.ComputeHeaderHash(header)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoBulkRemove(blockIndex, []string{hex.EncodeToString(headerHash)})
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header data.HeaderHandler, body *block.Body) error {
	encodedMiniblocksHashes := ei.miniblocksProc.GetMiniblocksHashesHexEncoded(header, body)
	if len(encodedMiniblocksHashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoBulkRemove(miniblocksIndex, encodedMiniblocksHashes)
}

// RemoveTransactions will remove transaction that are in miniblock from the elasticsearch server
func (ei *elasticProcessor) RemoveTransactions(header data.HeaderHandler, body *block.Body) error {
	encodedTxsHashes := ei.txProc.GetRewardsTxsHashesHexEncoded(header, body)
	if len(encodedTxsHashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoBulkRemove(txIndex, encodedTxsHashes)
}

// SetTxLogsProcessor will set tx logs processor
func (ei *elasticProcessor) SetTxLogsProcessor(txLogProcessor process.TransactionLogProcessorDatabase) {
	ei.txProc.SetTxLogsProcessor(txLogProcessor)
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error) {
	if !ei.isIndexEnabled(miniblocksIndex) {
		return map[string]bool{}, nil
	}

	mbs := ei.miniblocksProc.PrepareDBMiniblocks(header, body)
	if len(mbs) == 0 {
		return make(map[string]bool), nil
	}

	miniblocksInDBMap, err := ei.miniblocksInDBMap(mbs)
	if err != nil {
		log.Warn("elasticProcessor.SaveMiniblocks cannot get indexed miniblocks", "error", err)
	}

	buff := ei.miniblocksProc.SerializeBulkMiniBlocks(mbs, miniblocksInDBMap)
	return miniblocksInDBMap, ei.elasticClient.DoBulkRequest(buff, miniblocksIndex)
}

func (ei *elasticProcessor) miniblocksInDBMap(mbs []*types.Miniblock) (map[string]bool, error) {
	mbsHashes := make([]string, len(mbs))
	for idx := range mbs {
		mbsHashes[idx] = mbs[idx].Hash
	}

	return ei.getExistingObjMap(mbsHashes, miniblocksIndex)
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

	preparedResults := ei.txProc.PrepareTransactionsForDatabase(body, header, txPool)
	buffSlice, err := ei.txProc.SerializeTransactions(preparedResults.Transactions, selfShardID, mbsInDb)
	if err != nil {
		return err
	}

	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], txIndex)
		if err != nil {
			log.Warn("elasticProcessor.SaveTransactions cannot index bulk of transactions", "error", err)
			return err
		}
	}

	err = ei.indexScResults(preparedResults.ScResults)
	if err != nil {
		log.Warn("elasticProcessor.SaveTransactions cannot index bulk of smart contract results", "error", err)
	}

	err = ei.indexReceipts(preparedResults.Receipts)
	if err != nil {
		log.Warn("elasticProcessor.SaveTransactions cannot index bulk of receipts", "error", err)
	}

	return ei.indexAlteredAccounts(preparedResults.AlteredAccounts)
}

// SaveShardStatistics will prepare and save information about a shard statistics in elasticsearch server
func (ei *elasticProcessor) SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error {
	if !ei.isIndexEnabled(tpsIndex) {
		return nil
	}

	generalInfo, shardsInfo := ei.generalInfoProc.PrepareGeneralInfo(tpsBenchmark)
	buff := ei.generalInfoProc.SerializeGeneralInfo(generalInfo, shardsInfo, tpsIndex)

	return ei.elasticClient.DoBulkRequest(buff, tpsIndex)
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []types.ValidatorRatingInfo) error {
	if !ei.isIndexEnabled(ratingIndex) {
		return nil
	}

	buffSlice, err := ei.validatorsProc.SerializeValidatorsRating(index, validatorsRatingInfo)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], ratingIndex)
		if err != nil {
			log.Warn("elasticProcessor.SaveValidatorsRating cannot index validators rating", "error", err)
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

	validatorsPubKeys := ei.validatorsProc.PrepareValidatorsPublicKeys(shardValidatorsPubKeys)
	buff, err := ei.validatorsProc.SerializeValidatorsPubKeys(validatorsPubKeys)
	if err != nil {
		return err
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

	buff := ei.generalInfoProc.SerializeRoundsInfo(infos)

	return ei.elasticClient.DoBulkRequest(buff, roundIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(alteredAccounts map[string]*types.AlteredAccount) error {
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

func (ei *elasticProcessor) saveAccountsESDT(wrappedAccounts []*types.AccountESDT) error {
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
func (ei *elasticProcessor) SaveAccounts(accts []*types.AccountEGLD) error {
	if !ei.isIndexEnabled(accountsIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsMapEGLD(accts)
	err := ei.serializeAndIndexAccounts(accountsMap, accountsIndex, false)
	if err != nil {
		return err
	}

	return ei.saveAccountsHistory(accountsMap)
}

func (ei *elasticProcessor) serializeAndIndexAccounts(accountsMap map[string]*types.AccountInfo, index string, areESDTAccounts bool) error {
	buffSlice, err := ei.accountsProc.SerializeAccounts(accountsMap, areESDTAccounts)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], index)
		if err != nil {
			log.Warn("elasticProcessor.serializeAndIndexAccounts cannot index bulk of accounts",
				"index", index, "error", err)
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
	buffSlice, err := ei.accountsProc.SerializeAccountsHistory(accountsMap)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], index)
		if err != nil {
			log.Warn("elasticProcessor.serializeAndIndexAccountsHistory cannot index bulk of accounts history",
				"index", index, "error", err.Error())
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) indexScResults(scrs []*types.ScResult) error {
	if !ei.isIndexEnabled(scResultsIndex) {
		return nil
	}

	buffSlice, err := ei.txProc.SerializeScResults(scrs)
	if err != nil {
		return err
	}

	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], scResultsIndex)
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

	buffSlice, err := ei.txProc.SerializeReceipts(receipts)
	if err != nil {
		return err
	}

	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], receiptsIndex)
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
