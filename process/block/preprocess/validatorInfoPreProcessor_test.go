package preprocess

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorInfoPreprocessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewValidatorInfoPreprocessor(
		nil,
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewValidatorInfoPreprocessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()
	bytesss, _ := hex.DecodeString("a60e92147044fb2387d3430140d8b412140988b0221ebd0702e6714182326fd9")

	blSov := &block.SovereignChainHeader{}
	err := json.Unmarshal([]byte("{\"header\":{\"nonce\":27,\"prevHash\":\"qNVxicGPO+sTf4ysrpHBmcRipz79E/btcreIRq2QfdA=\",\"prevRandSeed\":\"X3X4s/SW2VUktT47qmd+UxUmM1bIk3whLckc//aK5PmbFQu5sET8BwIF9hWj1yUC\",\"randSeed\":\"UwSbXVizYm9/ADUKBMtKkKSRuISxYaCryDkQ8NzGGrDykrmkB54hMfl3a5yPSTwR\",\"shardID\":0,\"timeStamp\":1687257812,\"round\":27,\"epoch\":0,\"blockBodyType\":0,\"miniBlockHeaders\":null,\"peerChanges\":null,\"rootHash\":\"U2HjVT40oqV6C4JC4T312OOHZqtc92e+5CpjruL2OCQ=\",\"txCount\":0,\"chainID\":\"bG9jYWwtdGVzdG5ldA==\",\"softwareVersion\":\"UzE=\",\"accumulatedFees\":0,\"developerFees\":0},\"validatorStatsRootHash\":\"U2HjVT40oqV6C4JC4T312OOHZqtc92e+5CpjruL2OCQ=\",\"extendedShardHeaderHashes\":[\"pg6SFHBE+yOH00MBQNi0EhQJiLAiHr0HAuZxQYIyb9k=\"]}"),
		blSov)

	marshaller := &marshal.GogoProtoMarshalizer{}

	_, err = marshaller.Marshal(blSov)
	require.Nil(t, err)

	fmt.Println(len(bytesss))

	blSovPrev := &block.SovereignChainHeader{}
	err = json.Unmarshal([]byte("{\"header\":{\"nonce\":26,\"prevHash\":\"23pgKVNu2ni8HKYVDxlBuopiC2vUMqvWy6fHdPKnH1c=\",\"prevRandSeed\":\"bDkr3/pPmWaINqMVEGHT/4Ms1BOzdh9WLzDOSmfoPwW9rA/njsjbo7GUU9yvjukN\",\"randSeed\":\"X3X4s/SW2VUktT47qmd+UxUmM1bIk3whLckc//aK5PmbFQu5sET8BwIF9hWj1yUC\",\"shardID\":0,\"timeStamp\":1687257807,\"round\":26,\"epoch\":0,\"blockBodyType\":0,\"miniBlockHeaders\":[{\"hash\":\"aiKND1hCD/XGFCIqVqCbhaKmd5PjPE3zWDPQ3+5a004=\",\"senderShardID\":4294967293,\"receiverShardID\":4294967294,\"txCount\":1,\"type\":90}],\"peerChanges\":null,\"rootHash\":\"U2HjVT40oqV6C4JC4T312OOHZqtc92e+5CpjruL2OCQ=\",\"txCount\":0,\"chainID\":\"bG9jYWwtdGVzdG5ldA==\",\"softwareVersion\":\"UzE=\",\"accumulatedFees\":0,\"developerFees\":0},\"validatorStatsRootHash\":\"U2HjVT40oqV6C4JC4T312OOHZqtc92e+5CpjruL2OCQ=\",\"extendedShardHeaderHashes\":[\"pj1SR2G9mr0MEwhWTTDmuuiBrXBV+7Sfa+bYwv5VWkQ=\",\"eHmojq/i/ZArNkWK3rNFPJUosLmvlZWywbRu9vltCjw=\"]}"), blSovPrev)

	_, err = marshaller.Marshal(blSovPrev)
	require.Nil(t, err)

	newSovBl := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce:              0x11c,
			PrevHash:           []byte{0x97, 0x1c, 0x92, 0x75, 0x54, 0x2f, 0xe4, 0x44, 0x1a, 0xed, 0xce, 0xa7, 0xec, 0xa5, 0xeb, 0xf9, 0x94, 0x71, 0x81, 0xc9, 0x36, 0x43, 0xb7, 0xd, 0x75, 0xb6, 0x44, 0x73, 0xe1, 0xd9, 0x81, 0x2c},
			PrevRandSeed:       []byte{0x53, 0x2a, 0xb, 0x10, 0xc9, 0xbc, 0xf3, 0xad, 0xc6, 0x56, 0xe6, 0x3e, 0xde, 0x85, 0x37, 0xb8, 0x82, 0x3c, 0x8a, 0x9e, 0x2, 0xe4, 0x62, 0xa0, 0x11, 0x5c, 0x69, 0x44, 0xc1, 0x38, 0x65, 0xcb, 0xc1, 0x1a, 0x29, 0x1c, 0x61, 0x88, 0xf3, 0xe, 0xa3, 0xfc, 0x5b, 0x46, 0xf5, 0xad, 0xb6, 0x8},
			RandSeed:           []byte{0x31, 0xc2, 0xc6, 0xbb, 0xb6, 0xa, 0xf2, 0xbe, 0x9c, 0xa0, 0x2, 0xe8, 0x1e, 0x46, 0xb9, 0x55, 0x7b, 0xa2, 0x94, 0x52, 0xfb, 0x8, 0x1a, 0xac, 0x62, 0xa2, 0xb7, 0x43, 0x14, 0x95, 0xb4, 0x3b, 0x52, 0x10, 0xd5, 0x93, 0x98, 0x22, 0x9c, 0xab, 0xb9, 0xcc, 0x72, 0xe2, 0xe3, 0x6b, 0x31, 0x8e},
			PubKeysBitmap:      []byte(nil),
			ShardID:            0x0,
			TimeStamp:          0x649190a2,
			Round:              0x11c,
			Epoch:              0x0,
			BlockBodyType:      0,
			Signature:          []byte(nil),
			LeaderSignature:    []byte(nil),
			RootHash:           []byte{0x53, 0x61, 0xe3, 0x55, 0x3e, 0x34, 0xa2, 0xa5, 0x7a, 0xb, 0x82, 0x42, 0xe1, 0x3d, 0xf5, 0xd8, 0xe3, 0x87, 0x66, 0xab, 0x5c, 0xf7, 0x67, 0xbe, 0xe4, 0x2a, 0x63, 0xae, 0xe2, 0xf6, 0x38, 0x24},
			MetaBlockHashes:    [][]uint8(nil),
			TxCount:            0x0,
			EpochStartMetaHash: []byte(nil),
			ReceiptsHash:       []byte(nil),
			ChainID:            []byte{0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2d, 0x74, 0x65, 0x73, 0x74, 0x6e, 0x65, 0x74},
			SoftwareVersion:    []byte{0x53, 0x31},
			Reserved:           []byte(nil),
		},
		ValidatorStatsRootHash:    []byte{0x53, 0x61, 0xe3, 0x55, 0x3e, 0x34, 0xa2, 0xa5, 0x7a, 0xb, 0x82, 0x42, 0xe1, 0x3d, 0xf5, 0xd8, 0xe3, 0x87, 0x66, 0xab, 0x5c, 0xf7, 0x67, 0xbe, 0xe4, 0x2a, 0x63, 0xae, 0xe2, 0xf6, 0x38, 0x24},
		ExtendedShardHeaderHashes: [][]uint8{{0x41, 0xe5, 0xb6, 0xe2, 0x2, 0x75, 0xe1, 0x69, 0xa3, 0x13, 0x19, 0x7, 0x81, 0x7, 0xa3, 0xb2, 0xf5, 0xbc, 0x85, 0x3, 0xd5, 0xbd, 0xfe, 0x69, 0xf2, 0x81, 0x88, 0x1e, 0x6e, 0x12, 0xf9, 0x90}},
	}

	for i := 0; i < 1000; i++ {
		switch i % 3 {
		case 0:
			go func() {
				_, err = marshaller.Marshal(newSovBl.ShallowClone())
				require.Nil(t, err)
			}()
		case 1:
			go func() {
				newSovBl.SetLeaderSignature([]byte("321dasdas123421141242141231s123213123"))
			}()
		case 2:
			go func() {
				newSovBl.SetLeaderSignature([]byte(""))
			}()
		}

	}

	_, err = marshaller.Marshal(newSovBl)
	require.Nil(t, err)

	tdp := initDataPool()
	rtp, err := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		nil,
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewValidatorInfoPreprocessor_NilBlockSizeComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		nil,
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func TestNewValidatorInfoPreprocessor_NilValidatorInfoPoolShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		nil,
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilValidatorInfoPool, err)
}

func TestNewValidatorInfoPreprocessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		nil,
		&testscommon.EnableEpochsHandlerStub{},
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewValidatorInfoPreprocessor_NilEnableEpochHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		nil,
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewValidatorInfoPreprocessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, err := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, rtp)
}

func TestNewValidatorInfoPreprocessor_CreateMarshalizedDataShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	hash := make([][]byte, 0)
	res, err := rtp.CreateMarshalledData(hash)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(res))
}

func TestNewValidatorInfoPreprocessor_ProcessMiniBlockInvalidMiniBlockTypeShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	txHashes := make([][]byte, 0)
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Equal(t, process.ErrWrongTypeInMiniBlock, err)
}

func TestNewValidatorInfoPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	txHashes := make([][]byte, 0)
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Nil(t, err)
}

func TestNewValidatorInfoPreprocessor_ProcessMiniBlockNotFromMeta(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	txHashes := make([][]byte, 0)
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            block.PeerBlock,
	}

	preProcessorExecutionInfoHandlerMock := &testscommon.PreProcessorExecutionInfoHandlerMock{
		GetNumOfCrossInterMbsAndTxsCalled: getNumOfCrossInterMbsAndTxsZero,
	}

	_, _, _, err := rtp.ProcessMiniBlock(&mb1, haveTimeTrue, haveAdditionalTimeFalse, false, false, -1, preProcessorExecutionInfoHandlerMock)
	assert.Equal(t, process.ErrValidatorInfoMiniBlockNotFromMeta, err)
}

func TestNewValidatorInfoPreprocessor_RestorePeerBlockIntoPools(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &testscommon.MarshalizerMock{}
	blockSizeComputation := &testscommon.BlockSizeComputationStub{}

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		hasher,
		marshalizer,
		blockSizeComputation,
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1)
	miniBlockPool := testscommon.NewCacherMock()

	marshalizedMb, _ := marshalizer.Marshal(mb1)
	mbHash := hasher.Compute(string(marshalizedMb))

	foundMb, ok := miniBlockPool.Get(mbHash)
	assert.Nil(t, foundMb)
	assert.False(t, ok)

	numRestoredTxs, err := rtp.RestoreBlockDataIntoPools(blockBody, miniBlockPool)
	assert.Equal(t, 1, numRestoredTxs)
	assert.Nil(t, err)

	foundMb, ok = miniBlockPool.Get(mbHash)
	assert.NotNil(t, foundMb)
	assert.True(t, ok)
}

func TestNewValidatorInfoPreprocessor_RestoreOtherBlockTypeIntoPoolsShouldNotRestore(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &testscommon.MarshalizerMock{}
	blockSizeComputation := &testscommon.BlockSizeComputationStub{}

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		hasher,
		marshalizer,
		blockSizeComputation,
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.TxBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1)
	miniBlockPool := testscommon.NewCacherMock()

	marshalizedMb, _ := marshalizer.Marshal(mb1)
	mbHash := hasher.Compute(string(marshalizedMb))

	foundMb, ok := miniBlockPool.Get(mbHash)
	assert.Nil(t, foundMb)
	assert.False(t, ok)

	numRestoredTxs, err := rtp.RestoreBlockDataIntoPools(blockBody, miniBlockPool)
	assert.Equal(t, 0, numRestoredTxs)
	assert.Nil(t, err)

	foundMb, ok = miniBlockPool.Get(mbHash)
	assert.Nil(t, foundMb)
	assert.False(t, ok)
}

func TestNewValidatorInfoPreprocessor_RemovePeerBlockFromPool(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &testscommon.MarshalizerMock{}
	blockSizeComputation := &testscommon.BlockSizeComputationStub{}

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		hasher,
		marshalizer,
		blockSizeComputation,
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	marshalizedMb, _ := marshalizer.Marshal(mb1)
	mbHash := hasher.Compute(string(marshalizedMb))

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1)
	miniBlockPool := testscommon.NewCacherMock()
	miniBlockPool.Put(mbHash, marshalizedMb, len(marshalizedMb))

	foundMb, ok := miniBlockPool.Get(mbHash)
	assert.NotNil(t, foundMb)
	assert.True(t, ok)

	err := rtp.RemoveBlockDataFromPools(blockBody, miniBlockPool)
	assert.Nil(t, err)

	foundMb, ok = miniBlockPool.Get(mbHash)
	assert.Nil(t, foundMb)
	assert.False(t, ok)
}

func TestNewValidatorInfoPreprocessor_RemoveOtherBlockTypeFromPoolShouldNotRemove(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &testscommon.MarshalizerMock{}
	blockSizeComputation := &testscommon.BlockSizeComputationStub{}

	tdp := initDataPool()
	rtp, _ := NewValidatorInfoPreprocessor(
		hasher,
		marshalizer,
		blockSizeComputation,
		tdp.ValidatorsInfo(),
		genericMocks.NewChainStorerMock(0),
		&testscommon.EnableEpochsHandlerStub{},
	)

	txHashes := [][]byte{[]byte("tx_hash1")}
	mb1 := block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.TxBlock,
	}

	marshalizedMb, _ := marshalizer.Marshal(mb1)
	mbHash := hasher.Compute(string(marshalizedMb))

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1)
	miniBlockPool := testscommon.NewCacherMock()
	miniBlockPool.Put(mbHash, marshalizedMb, len(marshalizedMb))

	foundMb, ok := miniBlockPool.Get(mbHash)
	assert.NotNil(t, foundMb)
	assert.True(t, ok)

	err := rtp.RemoveBlockDataFromPools(blockBody, miniBlockPool)
	assert.Nil(t, err)

	foundMb, ok = miniBlockPool.Get(mbHash)
	assert.NotNil(t, foundMb)
	assert.True(t, ok)
}

func TestNewValidatorInfoPreprocessor_RestoreValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("restore validators info with not all txs found in storage", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("error")
		hasher := &hashingMocks.HasherMock{}
		marshalizer := &testscommon.MarshalizerMock{}
		blockSizeComputation := &testscommon.BlockSizeComputationStub{}
		storer := &storage.ChainStorerStub{
			GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
				return nil, expectedErr
			},
		}
		tdp := initDataPool()
		rtp, _ := NewValidatorInfoPreprocessor(
			hasher,
			marshalizer,
			blockSizeComputation,
			tdp.ValidatorsInfo(),
			storer,
			&testscommon.EnableEpochsHandlerStub{},
		)

		miniBlock := &block.MiniBlock{}
		err := rtp.restoreValidatorsInfo(miniBlock)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("restore validators info with all txs found in storage", func(t *testing.T) {
		t.Parallel()

		hasher := &hashingMocks.HasherMock{}
		marshalizer := &testscommon.MarshalizerMock{}
		blockSizeComputation := &testscommon.BlockSizeComputationStub{}
		shardValidatorInfoHash := []byte("hash")
		shardValidatorInfo := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}
		marshalledShardValidatorInfo, _ := marshalizer.Marshal(shardValidatorInfo)
		storer := &storage.ChainStorerStub{
			GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
				allShardValidatorsInfo := make(map[string][]byte)
				allShardValidatorsInfo[string(shardValidatorInfoHash)] = marshalledShardValidatorInfo
				return allShardValidatorsInfo, nil
			},
		}
		tdp := initDataPool()
		wasCalledWithExpectedKey := false
		tdp.ValidatorsInfoCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
					if bytes.Equal(key, shardValidatorInfoHash) {
						wasCalledWithExpectedKey = true
					}
				},
			}
		}
		rtp, _ := NewValidatorInfoPreprocessor(
			hasher,
			marshalizer,
			blockSizeComputation,
			tdp.ValidatorsInfo(),
			storer,
			&testscommon.EnableEpochsHandlerStub{},
		)

		miniBlock := &block.MiniBlock{}
		err := rtp.restoreValidatorsInfo(miniBlock)
		assert.Nil(t, err)
		assert.True(t, wasCalledWithExpectedKey)
	})
}

func TestValidatorInfoPreprocessor_SaveTxsToStorageShouldWork(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	txHash3 := []byte("txHash3")
	txHash4 := []byte("txHash4")

	tdp := initDataPool()

	tdp.ValidatorsInfoCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal(key, txHash1) {
					return nil, false
				}
				if bytes.Equal(key, txHash2) {
					return &state.ValidatorInfo{}, true
				}
				if bytes.Equal(key, txHash3) {
					return &state.ShardValidatorInfo{}, true
				}
				if bytes.Equal(key, txHash4) {
					return &rewardTx.RewardTx{}, true
				}
				return nil, false
			},
		}
	}

	putHashes := make([][]byte, 0)
	storer := &storage.ChainStorerStub{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			putHashes = append(putHashes, key)
			return nil
		},
	}

	vip, _ := NewValidatorInfoPreprocessor(
		&hashingMocks.HasherMock{},
		&testscommon.MarshalizerMock{},
		&testscommon.BlockSizeComputationStub{},
		tdp.ValidatorsInfo(),
		storer,
		&testscommon.EnableEpochsHandlerStub{},
	)

	err := vip.SaveTxsToStorage(nil)
	assert.Equal(t, process.ErrNilBlockBody, err)

	peersHashes := [][]byte{txHash1, txHash2, txHash3}
	rewardsHashes := [][]byte{txHash4}

	mb1 := block.MiniBlock{
		TxHashes: rewardsHashes,
		Type:     block.RewardsBlock,
	}
	mb2 := block.MiniBlock{
		TxHashes: peersHashes,
		Type:     block.PeerBlock,
	}

	blockBody := &block.Body{}
	blockBody.MiniBlocks = append(blockBody.MiniBlocks, &mb1, &mb2)

	err = vip.SaveTxsToStorage(blockBody)

	assert.Nil(t, err)
	require.Equal(t, 1, len(putHashes))
	assert.Equal(t, txHash3, putHashes[0])
}
