package core

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

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

// ToB64 encodes the given buff to base64
// This should be used only for display purposes!
func ToB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

// ToHex encodes the given buff to hex
// This should be used only for display purposes!
func ToHex(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return hex.EncodeToString(buff)
}

// CalculateHash marshalizes the interface and calculates its hash
func CalculateHash(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	object interface{},
) ([]byte, error) {
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
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

// GetShardIdString will return the string representation of the shard id
func GetShardIdString(shardId uint32) string {
	if shardId == math.MaxUint32 {
		return "metachain"
	}

	return fmt.Sprintf("%d", shardId)
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
