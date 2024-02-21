package main

import (
	"fmt"
	"log"

	abi "github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
)

func main() {
	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	// Create a struct to serialize.
	inputStruct := abi.StructValue{
		Fields: []abi.Field{
			{
				Value: abi.U8Value{Value: 1},
			},
			{
				Value: abi.U16Value{Value: 42},
			},
		},
	}

	encoded, err := serializer.Serialize([]any{inputStruct})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Encoded:", encoded)

	// Prepare an empty struct to deserialize into.
	outputStruct := abi.StructValue{
		Fields: []abi.Field{
			{
				Value: &abi.U8Value{},
			},
			{
				Value: &abi.U16Value{},
			},
		},
	}

	err = serializer.Deserialize(encoded, []any{&outputStruct})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoded:")
	fmt.Println(outputStruct.Fields[0].Value.(*abi.U8Value).Value)
	fmt.Println(outputStruct.Fields[1].Value.(*abi.U16Value).Value)
}
