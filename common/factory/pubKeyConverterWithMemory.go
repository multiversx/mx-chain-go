package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/storage/txcache/impl/maps"
)

type pubkeyConverterWithMemory struct {
	backingConverter       core.PubkeyConverter
	memoryEncodedByDecoded *maps.ConcurrentMap
	memoryDecodedByEncoded *maps.ConcurrentMap
}

func newPubkeyConverterWithMemory(backingConverter core.PubkeyConverter) *pubkeyConverterWithMemory {
	memoryEncodedByDecoded := maps.NewConcurrentMap(256)
	memoryDecodedByEncoded := maps.NewConcurrentMap(256)

	return &pubkeyConverterWithMemory{
		backingConverter:       backingConverter,
		memoryEncodedByDecoded: memoryEncodedByDecoded,
		memoryDecodedByEncoded: memoryDecodedByEncoded,
	}
}

func (p *pubkeyConverterWithMemory) Len() int {
	return p.backingConverter.Len()
}

func (p *pubkeyConverterWithMemory) Decode(humanReadable string) ([]byte, error) {
	decodedRemembered, ok := p.memoryDecodedByEncoded.Get(humanReadable)
	if ok {
		return decodedRemembered.([]byte), nil
	}

	decoded, err := p.backingConverter.Decode(humanReadable)
	if err != nil {
		return nil, err
	}

	p.memoryDecodedByEncoded.Set(humanReadable, decoded)
	return decoded, nil
}

func (p *pubkeyConverterWithMemory) Encode(pkBytes []byte) (string, error) {
	encodedRemembered, ok := p.memoryEncodedByDecoded.Get(string(pkBytes))
	if ok {
		return encodedRemembered.(string), nil
	}

	encoded, err := p.backingConverter.Encode(pkBytes)
	if err != nil {
		return "", err
	}

	p.memoryEncodedByDecoded.Set(string(pkBytes), encoded)
	return encoded, nil
}

func (p *pubkeyConverterWithMemory) SilentEncode(pkBytes []byte, log core.Logger) string {
	encodedRemembered, ok := p.memoryEncodedByDecoded.Get(string(pkBytes))
	if ok {
		return encodedRemembered.(string)
	}

	encoded := p.backingConverter.SilentEncode(pkBytes, log)
	p.memoryEncodedByDecoded.Set(string(pkBytes), encoded)
	return encoded
}

func (p *pubkeyConverterWithMemory) EncodeSlice(pkBytesSlice [][]byte) ([]string, error) {
	encodedSlice := make([]string, len(pkBytesSlice))

	for i, pkBytes := range pkBytesSlice {
		encoded, err := p.Encode(pkBytes)
		if err != nil {
			return nil, err
		}

		encodedSlice[i] = encoded
	}

	return encodedSlice, nil
}

func (p *pubkeyConverterWithMemory) IsInterfaceNil() bool {
	return p == nil
}
