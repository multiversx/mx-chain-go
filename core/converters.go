package core

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"

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
func ToB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

// ToHex encodes the given buff to hex
func ToHex(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return "0x" + hex.EncodeToString(buff)
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

// SecondsToHumanFormat transform seconds input in a human friendly format
func SecondsToHumanFormat(input int) string {
	numSecondsInAMinute := 60
	numMinutesInAHour := 60
	numSecondsInAHours := numSecondsInAMinute * numMinutesInAHour
	result := ""

	hours := math.Floor(float64(input) / float64(numSecondsInAMinute) / float64(numMinutesInAHour))
	seconds := input % (numSecondsInAHours)
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
