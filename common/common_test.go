package common_test

import (
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	commonErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/require"
)

var testFlag = core.EnableEpochFlag("test flag")

func TestIsValidRelayedTxV3(t *testing.T) {
	t.Parallel()

	scr := &smartContractResult.SmartContractResult{}
	require.False(t, common.IsValidRelayedTxV3(scr))
	require.False(t, common.IsRelayedTxV3(scr))

	notRelayedTxV3 := &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(100),
		RcvAddr:   []byte("receiver"),
		SndAddr:   []byte("sender0"),
		GasPrice:  100,
		GasLimit:  10,
		Signature: []byte("signature"),
	}
	require.False(t, common.IsValidRelayedTxV3(notRelayedTxV3))
	require.False(t, common.IsRelayedTxV3(notRelayedTxV3))

	invalidRelayedTxV3 := &transaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(100),
		RcvAddr:     []byte("receiver"),
		SndAddr:     []byte("sender0"),
		GasPrice:    100,
		GasLimit:    10,
		Signature:   []byte("signature"),
		RelayerAddr: []byte("relayer"),
	}
	require.False(t, common.IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, common.IsRelayedTxV3(invalidRelayedTxV3))

	invalidRelayedTxV3 = &transaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(100),
		RcvAddr:          []byte("receiver"),
		SndAddr:          []byte("sender0"),
		GasPrice:         100,
		GasLimit:         10,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	}
	require.False(t, common.IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, common.IsRelayedTxV3(invalidRelayedTxV3))

	relayedTxV3 := &transaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(100),
		RcvAddr:          []byte("receiver"),
		SndAddr:          []byte("sender1"),
		GasPrice:         100,
		GasLimit:         10,
		Signature:        []byte("signature"),
		RelayerAddr:      []byte("relayer"),
		RelayerSignature: []byte("signature"),
	}
	require.True(t, common.IsValidRelayedTxV3(relayedTxV3))
	require.True(t, common.IsRelayedTxV3(relayedTxV3))
}

func TestIsConsensusBitmapValid(t *testing.T) {
	t.Parallel()

	log := &testscommon.LoggerStub{}

	pubKeys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	t.Run("wrong size bitmap", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8)

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, false)
		require.Equal(t, common.ErrWrongSizeBitmap, err)
	})

	t.Run("not enough signatures", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8+1)
		bitmap[0] = 0x07

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, false)
		require.Equal(t, common.ErrNotEnoughSignatures, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8+1)
		bitmap[0] = 0x77
		bitmap[1] = 0x01

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, false)
		require.Nil(t, err)
	})

	t.Run("should work with fallback validation", func(t *testing.T) {
		t.Parallel()

		bitmap := make([]byte, len(pubKeys)/8+1)
		bitmap[0] = 0x77
		bitmap[1] = 0x01

		err := common.IsConsensusBitmapValid(log, pubKeys, bitmap, true)
		require.Nil(t, err)
	})
}

func TestIsEpochChangeBlockForFlagActivation(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(123)
	eeh := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			require.Equal(t, testFlag, flag)
			return providedEpoch
		},
	}

	epochStartHeaderSameEpoch := &block.HeaderV2{
		Header: &block.Header{
			EpochStartMetaHash: []byte("meta hash"),
			Epoch:              providedEpoch,
		},
	}
	notEpochStartHeaderSameEpoch := &block.HeaderV2{
		Header: &block.Header{
			Epoch: providedEpoch,
		},
	}
	epochStartHeaderOtherEpoch := &block.HeaderV2{
		Header: &block.Header{
			EpochStartMetaHash: []byte("meta hash"),
			Epoch:              providedEpoch + 1,
		},
	}
	notEpochStartHeaderOtherEpoch := &block.HeaderV2{
		Header: &block.Header{
			Epoch: providedEpoch + 1,
		},
	}

	require.True(t, common.IsEpochChangeBlockForFlagActivation(epochStartHeaderSameEpoch, eeh, testFlag))
	require.False(t, common.IsEpochChangeBlockForFlagActivation(notEpochStartHeaderSameEpoch, eeh, testFlag))
	require.False(t, common.IsEpochChangeBlockForFlagActivation(epochStartHeaderOtherEpoch, eeh, testFlag))
	require.False(t, common.IsEpochChangeBlockForFlagActivation(notEpochStartHeaderOtherEpoch, eeh, testFlag))
}

func TestGetShardIDs(t *testing.T) {
	t.Parallel()

	shardIDs := common.GetShardIDs(2)
	require.Equal(t, 3, len(shardIDs))
	_, hasShard0 := shardIDs[0]
	require.True(t, hasShard0)
	_, hasShard1 := shardIDs[1]
	require.True(t, hasShard1)
	_, hasShardM := shardIDs[core.MetachainShardId]
	require.True(t, hasShardM)
}

func TestGetBitmapSize(t *testing.T) {
	t.Parallel()

	require.Equal(t, 1, common.GetBitmapSize(8))
	require.Equal(t, 2, common.GetBitmapSize(8+1))
	require.Equal(t, 2, common.GetBitmapSize(8*2-1))
	require.Equal(t, 50, common.GetBitmapSize(8*50)) // 400 consensus size
}

func TestConsesusGroupSizeForShardAndEpoch(t *testing.T) {
	t.Parallel()

	t.Run("shard node", func(t *testing.T) {
		t.Parallel()

		groupSize := uint32(400)

		size := common.ConsensusGroupSizeForShardAndEpoch(
			&testscommon.LoggerStub{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						ShardConsensusGroupSize: groupSize,
					}, nil
				},
			},
			1,
			2,
		)

		require.Equal(t, int(groupSize), size)
	})

	t.Run("meta node", func(t *testing.T) {
		t.Parallel()

		groupSize := uint32(400)

		size := common.ConsensusGroupSizeForShardAndEpoch(
			&testscommon.LoggerStub{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						MetachainConsensusGroupSize: groupSize,
					}, nil
				},
			},
			core.MetachainShardId,
			2,
		)

		require.Equal(t, int(groupSize), size)
	})

	t.Run("on fail, use current parameters", func(t *testing.T) {
		t.Parallel()

		groupSize := uint32(400)

		size := common.ConsensusGroupSizeForShardAndEpoch(
			&testscommon.LoggerStub{},
			&chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{}, errors.New("fail")
				},
				CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
					return config.ChainParametersByEpochConfig{
						MetachainConsensusGroupSize: groupSize,
					}
				},
			},
			core.MetachainShardId,
			2,
		)

		require.Equal(t, int(groupSize), size)
	})
}

func TestGetHeaderTimestamps(t *testing.T) {
	t.Parallel()

	t.Run("nil checks", func(t *testing.T) {
		t.Parallel()

		header := &block.Header{
			Epoch:     2,
			TimeStamp: 123,
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag != common.SupernovaFlag
			},
		}

		_, _, err := common.GetHeaderTimestamps(nil, enableEpochsHandler)
		require.Equal(t, common.ErrNilHeaderHandler, err)

		_, _, err = common.GetHeaderTimestamps(header, nil)
		require.Equal(t, commonErrors.ErrNilEnableEpochsHandler, err)
	})

	t.Run("before supernova epoch activation", func(t *testing.T) {
		t.Parallel()

		header := &block.Header{
			Epoch:     2,
			TimeStamp: 123,
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag != common.SupernovaFlag
			},
		}

		timestampSec, timestampMs, _ := common.GetHeaderTimestamps(header, enableEpochsHandler)
		require.Equal(t, uint64(123), timestampSec)
		require.Equal(t, uint64(123000), timestampMs)
	})

	t.Run("after supernova epoch activation", func(t *testing.T) {
		t.Parallel()

		header := &block.Header{
			Epoch:     2,
			TimeStamp: 1234567, // as milliseconds
		}

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.SupernovaFlag
			},
		}

		timestampSec, timestampMs, _ := common.GetHeaderTimestamps(header, enableEpochsHandler)
		require.Equal(t, uint64(1234), timestampSec)
		require.Equal(t, uint64(1234567), timestampMs)
	})
}

// Structures for testing prettify functions
type inner struct {
	Bytes   []byte    `json:"bytes"`
	Big     *big.Int  `json:"big"`
	Float   big.Float `json:"float"`
	Rat     *big.Rat  `json:"rat"`
	private string    // unexported
}

type testStruct struct {
	InnerVal   inner        `json:"innerVal"`
	ByteArray  [4]byte      `json:"byteArray"`
	IntArray   []int        `json:"intArray"`
	FloatSlice []*big.Float `json:"floatSlice"`
	NilPtr     *big.Int     `json:"nilPtr"`
}

func TestPrettifyStruct(t *testing.T) {
	t.Parallel()

	t.Run("should return 'nil' for nil struct", func(t *testing.T) {
		t.Parallel()
		result, _ := common.PrettifyStruct(nil)
		require.Equal(t, "nil", result)
	})

	t.Run("with simple struct type", func(t *testing.T) {
		t.Parallel()

		type testStruct struct {
			Field1 string
			Field2 int
		}

		ts := &testStruct{
			Field1: "value1",
			Field2: 42,
		}

		expected := `{"Field1":"value1","Field2":42}`
		result, _ := common.PrettifyStruct(ts)
		require.Equal(t, expected, result)
	})

	t.Run("with array of simple struct type", func(t *testing.T) {
		t.Parallel()

		type testStruct struct {
			Field1 string
			Field2 int
		}

		ts := &testStruct{
			Field1: "value1",
			Field2: 42,
		}

		ts1 := &testStruct{
			Field1: "value2",
			Field2: 84,
		}
		tsArray := []*testStruct{ts, ts1}
		expected := `[{"Field1":"value1","Field2":42},{"Field1":"value2","Field2":84}]`
		result, _ := common.PrettifyStruct(tsArray)

		require.Equal(t, expected, result)
	})

	t.Run("with complex struct type", func(t *testing.T) {
		t.Parallel()
		v := testStruct{
			InnerVal: inner{
				Bytes:   []byte("some-bytes"),
				Big:     big.NewInt(42_000_000),
				Float:   *big.NewFloat(123.456),
				Rat:     big.NewRat(355, 113),
				private: "should-not-be-visible",
			},
			ByteArray:  [4]byte{'t', 'e', 's', 't'},
			IntArray:   []int{10, 20, 30},
			FloatSlice: []*big.Float{big.NewFloat(0.1), big.NewFloat(0.2)},
			NilPtr:     nil,
		}

		out, err := common.PrettifyStruct(v)
		require.NoError(t, err)
		expected := `{"byteArray":"74657374","floatSlice":["0.1","0.2"],"innerVal":{"big":"42000000","bytes":"736f6d652d6279746573","float":"123.456","private":"\u003cunexported\u003e","rat":"355/113"},"intArray":[10,20,30],"nilPtr":null}`
		require.Equal(t, expected, out)
	})

	t.Run("with minimal headers", func(t *testing.T) {
		t.Parallel()

		hdr := &block.Header{
			Nonce:            2,
			Round:            2,
			PrevHash:         []byte("prevHash"),
			PrevRandSeed:     []byte("prevRandSeed"),
			Signature:        []byte("signature"),
			PubKeysBitmap:    []byte("00110"),
			ShardID:          0,
			RootHash:         []byte("rootHash"),
			MiniBlockHeaders: []block.MiniBlockHeader{},
		}

		hdrv2 := &block.HeaderV2{
			Header:               hdr,
			ScheduledGasProvided: 0,
		}
		var h data.HeaderHandler
		h = hdrv2
		prettified, _ := common.PrettifyStruct(h)
		expected := `{"header":{"accumulatedFees":null,"blockBodyType":0,"chainID":"","developerFees":null,"epoch":0,"epochStartMetaHash":"","leaderSignature":"","metaBlockHashes":[],"miniBlockHeaders":[],"nonce":2,"peerChanges":[],"prevHash":"7072657648617368","prevRandSeed":"7072657652616e6453656564","pubKeysBitmap":"3030313130","randSeed":"","receiptsHash":"","reserved":"","rootHash":"726f6f7448617368","round":2,"shardID":0,"signature":"7369676e6174757265","softwareVersion":"","timeStamp":0,"txCount":0},"scheduledAccumulatedFees":null,"scheduledDeveloperFees":null,"scheduledGasPenalized":0,"scheduledGasProvided":0,"scheduledGasRefunded":0,"scheduledRootHash":""}`
		require.Equal(t, expected, prettified)

		metaHeader := &block.MetaBlock{
			Nonce:         2,
			Round:         2,
			PrevHash:      []byte("prevHash"),
			PrevRandSeed:  []byte("prevRandSeed"),
			Signature:     []byte("signature"),
			PubKeysBitmap: []byte("00110"),
			RootHash:      []byte("rootHash"),
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					TxCount: 100,
				},
			},
		}
		h = metaHeader
		prettified, _ = common.PrettifyStruct(h)
		expected = `{"accumulatedFees":null,"accumulatedFeesInEpoch":null,"chainID":"","devFeesInEpoch":null,"developerFees":null,"epoch":0,"epochStart":{"economics":{"nodePrice":null,"prevEpochStartHash":"","prevEpochStartRound":0,"rewardsForProtocolSustainability":null,"rewardsPerBlock":null,"totalNewlyMinted":null,"totalSupply":null,"totalToDistribute":null},"lastFinalizedHeaders":[]},"leaderSignature":"","miniBlockHeaders":[],"nonce":2,"peerInfo":[],"prevHash":"7072657648617368","prevRandSeed":"7072657652616e6453656564","pubKeysBitmap":"3030313130","randSeed":"","receiptsHash":"","reserved":"","rootHash":"726f6f7448617368","round":2,"shardInfo":[{"accumulatedFees":null,"developerFees":null,"epoch":0,"headerHash":"","lastIncludedMetaNonce":0,"nonce":0,"numPendingMiniBlocks":0,"prevHash":"","prevRandSeed":"","pubKeysBitmap":"","round":0,"shardID":0,"shardMiniBlockHeaders":[],"signature":"","txCount":100}],"signature":"7369676e6174757265","softwareVersion":"","timeStamp":0,"txCount":0,"validatorStatsRootHash":""}`
		require.Equal(t, expected, prettified)
	})

	t.Run("with complete headers", func(t *testing.T) {
		t.Parallel()

		headerV2 := `{"header":{"nonce":481,"prevHash":"nNqMnj/cTiZYVMq2WW8bh9vhiN69D/AIIm7wLn/nm+0=","prevRandSeed":"sG4G+2bvTXI/htmsnJAxfQXd9oTZe5KNZ5W766kCFBbFxl7B9yZ7YCiTyzkO5X2T","randSeed":"ACmbkLg73NmmZ8sGHFhe3MYtdK46wbKIUt+ivJsdHJld82HIAvnzjF0ezqUOdsOW","shardID":0,"timeStamp":1752768353,"round":0,"epoch":5,"blockBodyType":0,"leaderSignature":"aqchjZKetNFlMalYh+sgxcOdTxDABP7iIz5wOo5qirDI7BnhkuSQtDXhloNNZDqD","miniBlockHeaders":[{"hash":"JZ0x61Ds1sC7788L5sd3HM4CdcCsQxImNnwDY9sseUA=","senderShardID":4294967295,"receiverShardID":0,"txCount":5,"type":255,"reserved":"IAQ="},{"hash":"bvoV+3E/L3glspLncniBWc5Y913oZf9xLVdcbCdQ2R4=","senderShardID":4294967295,"receiverShardID":4294967280,"txCount":6,"type":60,"reserved":"IAU="},{"hash":"h4DQEYG9V7cCfWwC7n3HR7JRRXyxxP3wi/1aQYwaIWM=","senderShardID":4294967295,"receiverShardID":4294967280,"txCount":5,"type":60,"reserved":"IAQ="}],"peerChanges":null,"rootHash":"1vyoepMXFURwMyBaRW5bEfPVJnaD9R1VLo/w6edvSO8=","metaBlockHashes":["aZiUegDAPYMjxFoDZifKS1OtucE82KjcBcV7JOAnKrk="],"txCount":16,"epochStartMetaHash":"aZiUegDAPYMjxFoDZifKS1OtucE82KjcBcV7JOAnKrk=","receiptsHash":"DldRwCblQ7Loqy6wYJnaodHl30d3j3eH+qtFzfEv46g=","chainID":"bG9jYWxuZXQ=","softwareVersion":"Mg==","accumulatedFees":0,"developerFees":0},"scheduledRootHash":"SZRu3iHeUgmPfL99TQapZNOcqKSWYmp5rMrAuOMbmu0=","scheduledAccumulatedFees":0,"scheduledDeveloperFees":0,"scheduledGasProvided":0,"scheduledGasPenalized":0,"scheduledGasRefunded":0}`
		metablock := `{"nonce":184,"epoch":2,"round":0,"timeStamp":1752766553,"shardInfo":[{"headerHash":"N4Be23RIX4Hdb/IX8N9Rn9IVrDwNv0x/aRBG3DeZ59s=","shardMiniBlockHeaders":[{"hash":"SQnnrD2Cv9UbULqY2vrdsP9pKzVrp9lUgMaKf8N/VQs=","senderShardID":0,"receiverShardID":0,"txCount":100,"type":0}],"prevRandSeed":"AtTtjVgLLCR1vcN5lhMgKAXSQ+uGgQJQAGCIRXpRur2WgyOFWVwGsvB0XNr5tT2D","round":205,"prevHash":"5DzInuk8HiY/x21RCIAaLnmEp2pNcj3GFhjV/D0ugeo=","nonce":182,"accumulatedFees":5000000000000000,"developerFees":0,"numPendingMiniBlocks":1,"lastIncludedMetaNonce":181,"shardID":0,"txCount":100,"epoch":2}],"peerInfo":null,"leaderSignature":"MaAFUyniShBNVbL01Mf5WJOAh0ypTKcjFtQ4E+wODRWpUWjb1/icT07eeEK5n7oT","prevHash":"p5RrqnclvenWpggjZazqDuNMSh/BAKXUUZOW4Ty3R80=","prevRandSeed":"n+jWtdpAJrz1G8YyxNtn6aKuMSwrpVhwzhaHVbEsTIJe0i5N3gzl73QWxaHWjUiU","randSeed":"ek0OGMLItkHOwp/AtNtM8jup4ZKUgXw2xPpMEvARqWUo+xiMai7K0Zt5n/EKl0QL","rootHash":"SA2azL3/LsUqsfofORRsha0dXjBlshUEVJELa9uBTuQ=","validatorStatsRootHash":"YWE259eQYLgZ94BeQE5Ur9/IuGlrQj3K3NMb/FnsYus=","miniBlockHeaders":null,"receiptsHash":"DldRwCblQ7Loqy6wYJnaodHl30d3j3eH+qtFzfEv46g=","epochStart":{"lastFinalizedHeaders":null,"economics":{"prevEpochStartRound":0}},"chainID":"bG9jYWxuZXQ=","softwareVersion":"Mg==","accumulatedFees":0,"accumulatedFeesInEpoch":5050000000000000,"developerFees":0,"devFeesInEpoch":0,"txCount":100}`

		header := &block.HeaderV2{}
		require.NoError(t, json.Unmarshal([]byte(headerV2), header))
		prettified, err := common.PrettifyStruct(header)
		require.NoError(t, err)
		t.Log("HeaderV2", prettified)

		meta := &block.MetaBlock{}
		require.NoError(t, json.Unmarshal([]byte(metablock), meta))
		prettified, err = common.PrettifyStruct(meta)
		require.NoError(t, err)
		t.Log("MetaBlock", prettified)

	})
}

func TestGetLastBaseExecutionResultHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil header, should return error", func(t *testing.T) {
		t.Parallel()

		var header data.HeaderHandler
		result, err := common.GetLastBaseExecutionResultHandler(header)
		require.Nil(t, result)
		require.Equal(t, common.ErrNilHeaderHandler, err)
	})
	t.Run("nil last execution result (wrong header), should return error", func(t *testing.T) {
		t.Parallel()

		result, err := common.GetLastBaseExecutionResultHandler(&block.Header{})
		require.Nil(t, result)
		require.Equal(t, common.ErrNilLastExecutionResultHandler, err)
	})
	t.Run("valid LastMetaExecutionResultHandler, should return handler", func(t *testing.T) {
		t.Parallel()

		baseMetaExecutionResultsHandler := &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash"),
				HeaderNonce: 100,
				HeaderRound: 200,
				RootHash:    []byte("rootHash"),
			},
		}

		header := &block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				NotarizedInRound: 201,
				ExecutionResult:  baseMetaExecutionResultsHandler,
			},
		}

		result, err := common.GetLastBaseExecutionResultHandler(header)
		require.NotNil(t, result)
		require.Nil(t, err)
		require.Equal(t, baseMetaExecutionResultsHandler, result)
	})
	t.Run("nil internal BaseMetaExecutionResultHandler, should return error", func(t *testing.T) {
		t.Parallel()

		header := &block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				NotarizedInRound: 201,
				ExecutionResult:  nil,
			},
		}

		result, err := common.GetLastBaseExecutionResultHandler(header)
		require.Nil(t, result)
		require.Equal(t, common.ErrNilBaseExecutionResult, err)
	})
	t.Run("valid LastShardExecutionResultHandler, should return handler", func(t *testing.T) {
		t.Parallel()

		baseExecutionResults := &block.BaseExecutionResult{
			HeaderHash:  []byte("hash"),
			HeaderNonce: 100,
			HeaderRound: 200,
			RootHash:    []byte("rootHash"),
		}
		header := &block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 201,
				ExecutionResult:  baseExecutionResults,
			},
		}

		result, err := common.GetLastBaseExecutionResultHandler(header)
		require.NotNil(t, result)
		require.Nil(t, err)
		require.Equal(t, baseExecutionResults, result)
	})

	t.Run("nil base execution result, should return error", func(t *testing.T) {
		t.Parallel()

		var baseExecutionResultsHandler *block.BaseExecutionResult
		header := &block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 201,
				ExecutionResult:  baseExecutionResultsHandler,
			},
		}

		result, err := common.GetLastBaseExecutionResultHandler(header)
		require.Nil(t, result)
		require.Equal(t, common.ErrNilBaseExecutionResult, err)
	})
}

func TestGetMiniBlockHeaderHandlersFromExecResults(t *testing.T) {
	t.Parallel()

	t.Run("should fail if nil base execution result", func(t *testing.T) {
		t.Parallel()

		retExecResult, err := common.GetMiniBlocksHeaderHandlersFromExecResult(nil, 1)
		require.Equal(t, common.ErrNilBaseExecutionResult, err)
		require.Nil(t, retExecResult)
	})

	t.Run("should fail if wrong type for meta", func(t *testing.T) {
		t.Parallel()

		execResult := &block.ExecutionResult{}

		retExecResult, err := common.GetMiniBlocksHeaderHandlersFromExecResult(execResult, core.MetachainShardId)
		require.Equal(t, common.ErrWrongTypeAssertion, err)
		require.Nil(t, retExecResult)
	})

	t.Run("should fail if wrong type for shard", func(t *testing.T) {
		t.Parallel()

		execResult := &block.BaseExecutionResult{}

		retExecResult, err := common.GetMiniBlocksHeaderHandlersFromExecResult(execResult, 1)
		require.Equal(t, common.ErrWrongTypeAssertion, err)
		require.Nil(t, retExecResult)
	})

	t.Run("should work for shard", func(t *testing.T) {
		t.Parallel()

		mbh1 := block.MiniBlockHeader{
			Hash: []byte("hash1"),
		}
		mbh2 := block.MiniBlockHeader{
			Hash: []byte("hash1"),
		}

		miniBlockHeaders := []block.MiniBlockHeader{
			mbh1,
			mbh2,
		}

		execResult := &block.ExecutionResult{
			MiniBlockHeaders: miniBlockHeaders,
		}

		expMiniBlockHandlers := []data.MiniBlockHeaderHandler{
			&mbh1,
			&mbh2,
		}

		retExecResult, err := common.GetMiniBlocksHeaderHandlersFromExecResult(execResult, 1)
		require.Nil(t, err)
		require.Equal(t, expMiniBlockHandlers, retExecResult)
	})

	t.Run("should work for meta", func(t *testing.T) {
		t.Parallel()

		mbh1 := block.MiniBlockHeader{
			Hash: []byte("hash1"),
		}
		mbh2 := block.MiniBlockHeader{
			Hash: []byte("hash1"),
		}

		miniBlockHeaders := []block.MiniBlockHeader{
			mbh1,
			mbh2,
		}

		execResult := &block.MetaExecutionResult{
			MiniBlockHeaders: miniBlockHeaders,
		}

		expMiniBlockHandlers := []data.MiniBlockHeaderHandler{
			&mbh1,
			&mbh2,
		}

		retExecResult, err := common.GetMiniBlocksHeaderHandlersFromExecResult(execResult, core.MetachainShardId)
		require.Nil(t, err)
		require.Equal(t, expMiniBlockHandlers, retExecResult)
	})
}

func TestPrepareLogEventsKey(t *testing.T) {
	t.Parallel()

	logs := common.PrepareLogEventsKey([]byte("LogsX"))
	require.Equal(t, "logsLogsX", string(logs))
}
