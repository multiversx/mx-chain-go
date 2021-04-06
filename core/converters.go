package core

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// ConvertBytes converts the input bytes in a readable string using multipliers (k, M, G)
func ConvertBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024.0)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024.0/1024.0)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/1024.0/1024.0/1024.0)
}

// CalculateHash marshalizes the interface and calculates its hash
func CalculateHash(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	object interface{},
) ([]byte, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	mrsData, err := marshalizer.Marshal(object)
	if err != nil {
		return nil, err
	}

	hash := hasher.Compute(string(mrsData))
	return hash, nil
}

func plural(count int, singular string) (result string) {
	if count < 2 {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}
	return
}

// SecondsToHourMinSec transform seconds input in a human friendly format
func SecondsToHourMinSec(input int) string {
	numSecondsInAMinute := 60
	numMinutesInAHour := 60
	numSecondsInAHour := numSecondsInAMinute * numMinutesInAHour
	result := ""

	hours := math.Floor(float64(input) / float64(numSecondsInAMinute) / float64(numMinutesInAHour))
	seconds := input % (numSecondsInAHour)
	minutes := math.Floor(float64(seconds) / float64(numSecondsInAMinute))
	seconds = input % numSecondsInAMinute

	if hours > 0 {
		result = plural(int(hours), "hour")
	}
	if minutes > 0 {
		result += plural(int(minutes), "minute")
	}
	if seconds > 0 {
		result += plural(seconds, "second")
	}

	return result
}

// GetShardIDString will return the string representation of the shard id
func GetShardIDString(shardID uint32) string {
	if shardID == math.MaxUint32 {
		return "metachain"
	}

	return fmt.Sprintf("%d", shardID)
}

// ConvertShardIDToUint32 converts shard id from string to uint32
func ConvertShardIDToUint32(shardIDStr string) (uint32, error) {
	if shardIDStr == "metachain" {
		return MetachainShardId, nil
	}

	shardID, err := strconv.ParseInt(shardIDStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint32(shardID), nil
}

// EpochStartIdentifier returns the string for the epoch start identifier
func EpochStartIdentifier(epoch uint32) string {
	return fmt.Sprintf("epochStartBlock_%d", epoch)
}

// IsUnknownEpochIdentifier return if the epoch identifier represents unknown epoch
func IsUnknownEpochIdentifier(identifier []byte) (bool, error) {
	splitString := strings.Split(string(identifier), "_")
	if len(splitString) < 2 || len(splitString[1]) == 0 {
		return false, ErrInvalidIdentifierForEpochStartBlockRequest
	}

	epoch, err := strconv.ParseUint(splitString[1], 10, 32)
	if err != nil {
		return false, ErrInvalidIdentifierForEpochStartBlockRequest
	}

	if epoch == math.MaxUint32 {
		return true, nil
	}

	return false, nil
}

// CommunicationIdentifierBetweenShards is used to generate the identifier between shardID1 and shardID2
// identifier is generated such as the first shard from identifier is always smaller or equal than the last
func CommunicationIdentifierBetweenShards(shardId1 uint32, shardId2 uint32) string {
	if shardId1 == AllShardId || shardId2 == AllShardId {
		return ShardIdToString(AllShardId)
	}

	if shardId1 == shardId2 {
		return ShardIdToString(shardId1)
	}

	if shardId1 < shardId2 {
		return ShardIdToString(shardId1) + ShardIdToString(shardId2)
	}

	return ShardIdToString(shardId2) + ShardIdToString(shardId1)
}

// ShardIdToString returns the string according to the shard id
func ShardIdToString(shardId uint32) string {
	if shardId == MetachainShardId {
		return "_META"
	}
	if shardId == AllShardId {
		return "_ALL"
	}
	return fmt.Sprintf("_%d", shardId)
}

// ConvertToEvenHex converts the provided value in a hex string, even number of characters
func ConvertToEvenHex(value int) string {
	str := fmt.Sprintf("%x", value)
	if len(str)%2 != 0 {
		str = "0" + str
	}

	return str
}

// ConvertToEvenHexBigInt converts the provided value in a hex string, even number of characters
func ConvertToEvenHexBigInt(value *big.Int) string {
	str := value.Text(16)
	if len(str)%2 != 0 {
		str = "0" + str
	}

	return str
}

// ProcessDestinationShardAsObserver returns the shardID given the destination as observer string
func ProcessDestinationShardAsObserver(destinationShardIdAsObserver string) (uint32, error) {
	destShard := strings.ToLower(destinationShardIdAsObserver)
	if len(destShard) == 0 {
		return 0, fmt.Errorf("option DestinationShardAsObserver is not set in prefs.toml")
	}

	if destShard == NotSetDestinationShardID {
		return DisabledShardIDAsObserver, nil
	}

	if destShard == MetachainShardName {
		return MetachainShardId, nil
	}

	val, err := strconv.ParseUint(destShard, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("error parsing DestinationShardAsObserver option: " + err.Error())
	}

	return uint32(val), err
}

// AssignShardForPubKeyWhenNotSpecified will return the same shard ID when the same seed source is set to the randomizer
// This function sums all the bytes from the public key and based on a modulo operation it will return a shard ID
func AssignShardForPubKeyWhenNotSpecified(pubKey []byte, numShards uint32) uint32 {
	if len(pubKey) == 0 {
		return 0 // should never happen
	}

	lastByte := pubKey[len(pubKey)-1]
	numShardsIncludingMeta := numShards + 1

	randomShardID := uint32(lastByte) % numShardsIncludingMeta
	if randomShardID == numShards {
		randomShardID = MetachainShardId
	}

	return randomShardID
}
