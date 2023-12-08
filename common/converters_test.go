package common_test

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateHash_NilMarshalizer(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	hash, err := core.CalculateHash(nil, &hashingMocks.HasherMock{}, obj)
	assert.Nil(t, hash)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestCalculateHash_NilHasher(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	hash, err := core.CalculateHash(&mock.MarshalizerMock{}, nil, obj)
	assert.Nil(t, hash)
	assert.Equal(t, core.ErrNilHasher, err)
}

func TestCalculateHash_ErrMarshalizer(t *testing.T) {
	t.Parallel()

	obj := &batch.Batch{Data: [][]byte{[]byte("object")}}
	marshalizer := &mock.MarshalizerMock{
		Fail: true,
	}
	hash, err := core.CalculateHash(marshalizer, &hashingMocks.HasherMock{}, obj)
	assert.Nil(t, hash)
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestCalculateHash_NilObject(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hash, err := core.CalculateHash(marshalizer, &hashingMocks.HasherMock{}, nil)
	assert.Nil(t, hash)
	assert.Equal(t, mock.ErrNilObjectToMarshal, err)
}

func TestGetShardIdString(t *testing.T) {
	t.Parallel()

	shardIdMeta := uint32(math.MaxUint32)
	assert.Equal(t, "metachain", core.GetShardIDString(shardIdMeta))

	shardId37 := uint32(37)
	assert.Equal(t, "37", core.GetShardIDString(shardId37))
}

func TestEpochStartIdentifier(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	res := core.EpochStartIdentifier(epoch)
	assert.True(t, strings.Contains(res, fmt.Sprintf("%d", epoch)))
}

func TestIsUnknownEpochIdentifier_InvalidIdentifierShouldReturnTrue(t *testing.T) {
	t.Parallel()

	identifier := "epoch5"
	res, err := core.IsUnknownEpochIdentifier([]byte(identifier))
	assert.False(t, res)
	assert.Equal(t, core.ErrInvalidIdentifierForEpochStartBlockRequest, err)
}

func TestIsUnknownEpochIdentifier_IdentifierNotNumericShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier := "epochStartBlock_xx"
	res, err := core.IsUnknownEpochIdentifier([]byte(identifier))
	assert.False(t, res)
	assert.Equal(t, core.ErrInvalidIdentifierForEpochStartBlockRequest, err)
}

func TestIsUnknownEpochIdentifier_OkIdentifierShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier := "epochStartBlock_5"
	res, err := core.IsUnknownEpochIdentifier([]byte(identifier))
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestIsUnknownEpochIdentifier_Ok(t *testing.T) {
	t.Parallel()

	identifier := core.EpochStartIdentifier(math.MaxUint32)
	res, err := core.IsUnknownEpochIdentifier([]byte(identifier))
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestCalculateHash_Good(t *testing.T) {
	t.Parallel()

	initialObject := "random string"
	marshalCalled := false
	hashCalled := false
	marshaledData := "marshalized random string"
	hashedData := "hashed marshalized random string"
	hash, err := core.CalculateHash(
		&marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				marshalCalled = true
				assert.Equal(t, initialObject, obj)

				return []byte(marshaledData), nil
			},
		},
		&testscommon.HasherStub{
			ComputeCalled: func(s string) []byte {
				hashCalled = true
				assert.Equal(t, marshaledData, s)
				return []byte(hashedData)
			},
		},
		initialObject)

	assert.Nil(t, err)
	assert.True(t, marshalCalled)
	assert.True(t, hashCalled)
	assert.Equal(t, hashedData, string(hash))
}

func TestSecondsToHourMinSec_ShouldWork(t *testing.T) {
	t.Parallel()

	second := 1
	secondsInAMinute := 60
	secondsInAnHour := 3600

	testInputOutput := map[int]string{
		-1:                                   "",
		0:                                    "",
		second:                               "1 second ",
		2 * second:                           "2 seconds ",
		1 * secondsInAMinute:                 "1 minute ",
		1*secondsInAMinute + second:          "1 minute 1 second ",
		1*secondsInAMinute + 2*second:        "1 minute 2 seconds ",
		2*secondsInAMinute + second:          "2 minutes 1 second ",
		2*secondsInAMinute + 10*second:       "2 minutes 10 seconds ",
		59*secondsInAMinute + 59*second:      "59 minutes 59 seconds ",
		secondsInAnHour:                      "1 hour ",
		secondsInAnHour + second:             "1 hour 1 second ",
		secondsInAnHour + 2*second:           "1 hour 2 seconds ",
		secondsInAnHour + secondsInAMinute:   "1 hour 1 minute ",
		secondsInAnHour + 2*secondsInAMinute: "1 hour 2 minutes ",
		secondsInAnHour + secondsInAMinute + second:          "1 hour 1 minute 1 second ",
		secondsInAnHour + 2*secondsInAMinute + second:        "1 hour 2 minutes 1 second ",
		secondsInAnHour + 2*secondsInAMinute + 10*second:     "1 hour 2 minutes 10 seconds ",
		2*secondsInAnHour + 2*secondsInAMinute + 10*second:   "2 hours 2 minutes 10 seconds ",
		60*secondsInAnHour + 15*secondsInAMinute + 20*second: "60 hours 15 minutes 20 seconds ",
	}

	for input, expectedOutput := range testInputOutput {
		result := core.SecondsToHourMinSec(input)
		assert.Equal(t, expectedOutput, result)
	}
}

func TestCommunicationIdentifierBetweenShards(t *testing.T) {
	//print some shard identifiers and check that they match the current defined pattern

	for shard1 := uint32(0); shard1 < 5; shard1++ {
		for shard2 := uint32(0); shard2 < 5; shard2++ {
			identifier := core.CommunicationIdentifierBetweenShards(shard1, shard2)
			fmt.Printf("Shard1: %d, Shard2: %d, identifier: %s\n", shard1, shard2, identifier)

			if shard1 == shard2 {
				assert.Equal(t, fmt.Sprintf("_%d", shard1), identifier)
				continue
			}

			if shard1 < shard2 {
				assert.Equal(t, fmt.Sprintf("_%d_%d", shard1, shard2), identifier)
				continue
			}

			assert.Equal(t, fmt.Sprintf("_%d_%d", shard2, shard1), identifier)
		}
	}
}

func TestCommunicationIdentifierBetweenShards_Metachain(t *testing.T) {
	//print some shard identifiers and check that they match the current defined pattern

	assert.Equal(t, "_0_META", core.CommunicationIdentifierBetweenShards(0, core.MetachainShardId))
	assert.Equal(t, "_1_META", core.CommunicationIdentifierBetweenShards(core.MetachainShardId, 1))
	assert.Equal(t, "_META", core.CommunicationIdentifierBetweenShards(
		core.MetachainShardId,
		core.MetachainShardId,
	))
}

func TestConvertToEvenHex(t *testing.T) {
	t.Parallel()

	numCompares := 100000
	for i := 0; i < numCompares; i++ {
		str := core.ConvertToEvenHex(i)

		assert.True(t, len(str)%2 == 0)
		recovered, err := strconv.ParseInt(str, 16, 32)
		assert.Nil(t, err)
		assert.Equal(t, i, int(recovered))
	}
}

func TestConvertToEvenHexBigInt(t *testing.T) {
	t.Parallel()

	numCompares := 100000
	for i := 0; i < numCompares; i++ {
		bigInt := big.NewInt(int64(i))
		str := core.ConvertToEvenHexBigInt(bigInt)

		assert.True(t, len(str)%2 == 0)
		recovered, err := strconv.ParseInt(str, 16, 32)
		assert.Nil(t, err, str)
		assert.Equal(t, i, int(recovered))
	}
}

func TestConvertShardIDToUint32(t *testing.T) {
	t.Parallel()

	shardID, err := core.ConvertShardIDToUint32("metachain")
	assert.NoError(t, err)
	assert.Equal(t, core.MetachainShardId, shardID)

	id := uint32(0)
	shardIDStr := fmt.Sprintf("%d", id)
	shardID, err = core.ConvertShardIDToUint32(shardIDStr)
	assert.NoError(t, err)
	assert.Equal(t, id, shardID)

	shardID, err = core.ConvertShardIDToUint32("wrongID")
	assert.Error(t, err)
	assert.Equal(t, uint32(0), shardID)
}

func TestAssignShardForPubKeyWhenNotSpecified(t *testing.T) {
	t.Parallel()

	numShards := uint32(3)

	key := []byte{5, 7, 4} // 4 % 4 = 0
	require.Equal(t, uint32(0), common.AssignShardForPubKeyWhenNotSpecified(key, numShards))

	key = []byte{5, 7, 5} // 5 % 4 = 1
	require.Equal(t, uint32(1), common.AssignShardForPubKeyWhenNotSpecified(key, numShards))

	key = []byte{5, 7, 6} // 6 % 4 = 2
	require.Equal(t, uint32(2), common.AssignShardForPubKeyWhenNotSpecified(key, numShards))

	key = []byte{5, 7, 7} // 7 % 4 = 3 => metachain
	require.Equal(t, core.MetachainShardId, common.AssignShardForPubKeyWhenNotSpecified(key, numShards))

	key = []byte{5, 7, 8} // 8 % 4 = 0
	require.Equal(t, uint32(0), common.AssignShardForPubKeyWhenNotSpecified(key, numShards))

	key = []byte{} // empty
	require.Equal(t, uint32(0), common.AssignShardForPubKeyWhenNotSpecified(key, numShards))
}

func TestAssignShardForPubKeyWhenNotSpecifiedShouldReturnSameShardForSameKey(t *testing.T) {
	t.Parallel()

	key := []byte("test pub key number 0")
	numShards := uint32(3)

	result0 := common.AssignShardForPubKeyWhenNotSpecified(key, numShards)
	result1 := common.AssignShardForPubKeyWhenNotSpecified(key, numShards)

	require.Equal(t, result0, result1)
}

func TestShardAssignment(t *testing.T) {
	t.Skip()

	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	generatePubKey := func() []byte {
		_, pk := keyGen.GeneratePair()
		pkB, _ := pk.ToByteArray()

		return pkB
	}

	numShards := uint32(3)
	counts := map[uint32]uint64{
		0:                     0,
		1:                     0,
		2:                     0,
		core.MetachainShardId: 0,
	}

	numAccounts := 1000

	for i := 0; i < numAccounts; i++ {
		pubKey := generatePubKey()
		shId := common.AssignShardForPubKeyWhenNotSpecified(pubKey, numShards)
		counts[shId]++
	}

	for sh, cnt := range counts {
		fmt.Printf("Shard %d:\n\t\t%d accounts\n", sh, cnt)
	}
}

func TestProcessDestinationShardAsObserver(t *testing.T) {
	t.Parallel()

	t.Run("empty shard should error", testProcessDestinationShardAsObserver("", 0, "option DestinationShardAsObserver is not set in prefs.toml"))
	t.Run("disabled shard should return disabled",
		testProcessDestinationShardAsObserver(common.NotSetDestinationShardID, common.DisabledShardIDAsObserver, ""))
	t.Run("metachain should return metachain",
		testProcessDestinationShardAsObserver("metachain", common.MetachainShardId, ""))
	t.Run("MeTaChAiN should return metachain",
		testProcessDestinationShardAsObserver("MeTaChAiN", common.MetachainShardId, ""))
	t.Run("METACHAIN should return metachain",
		testProcessDestinationShardAsObserver("METACHAIN", common.MetachainShardId, ""))
	t.Run("invalid uint should error",
		testProcessDestinationShardAsObserver("not uint", 0, "error parsing DestinationShardAsObserver"))
	t.Run("should work",
		testProcessDestinationShardAsObserver("1", 1, ""))
}

func testProcessDestinationShardAsObserver(providedShard string, expectedShard uint32, expectedError string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		shard, err := common.ProcessDestinationShardAsObserver(providedShard)
		if len(expectedError) == 0 {
			require.NoError(t, err)
		} else {
			require.True(t, strings.Contains(err.Error(), expectedError))
		}
		require.Equal(t, expectedShard, shard)
	}
}

func TestSuffixedMetric(t *testing.T) {
	t.Parallel()

	providedMetric := common.MetricP2PPeerInfo
	providedSuffix := ""
	expectedMetric := providedMetric
	require.Equal(t, expectedMetric, common.SuffixedMetric(providedMetric, providedSuffix))

	providedSuffix = common.FullArchiveMetricSuffix
	expectedMetric = providedMetric + providedSuffix
	require.Equal(t, expectedMetric, common.SuffixedMetric(providedMetric, providedSuffix))
}
