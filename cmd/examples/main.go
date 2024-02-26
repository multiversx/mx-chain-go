package main

import (
	"fmt"
	"log"

	abi "github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
)

func main() {
	doBasicExample()
	doExampleWithTransferData()
}

func doBasicExample() {
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

func doExampleWithTransferData() {
	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	transferData := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: &abi.U64Value{},
			},
			{
				Name: "function",
				Value: &abi.OptionValue{
					Value: &abi.StringValue{},
				},
			},
			{
				Name: "args",
				Value: &abi.OptionValue{
					Value: &abi.OutputListValue{
						ItemCreator: func() any { return &abi.StringValue{} },
					},
				},
			},
			{
				Name: "gas_limit",
				Value: &abi.OptionValue{
					Value: &abi.U64Value{},
				},
			},
		},
	}

	err := serializer.Deserialize("00000000000000030001000000010000000361626300", []any{&transferData})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoded:")
	fmt.Println(transferData)
	fmt.Println("Nonce", transferData.Fields[0].Value.(*abi.U64Value).Value)
	fmt.Println("Args", transferData.Fields[2].Value.(*abi.OptionValue).Value.(*abi.OutputListValue).Items[0].(*abi.StringValue).Value)
}
