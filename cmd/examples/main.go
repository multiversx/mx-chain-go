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
						ItemCreator: func() any { return &abi.BytesValue{} },
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

	err := serializer.Deserialize("000000000000001a01000000076465706f7369740100000001000000200000000000000000050008b94c89df86fc204ed9d289f81a4b72fbe00a8fe7f6010000000001312d00", []any{&transferData})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoded:")
	fmt.Println(transferData)
	fmt.Println("Nonce", transferData.Fields[0].Value.(*abi.U64Value).Value)
	fmt.Println("Function", transferData.Fields[1].Value.(*abi.OptionValue).Value.(*abi.StringValue).Value)
	fmt.Println("Args", transferData.Fields[2].Value.(*abi.OptionValue).Value.(*abi.OutputListValue).Items[0].(*abi.BytesValue).Value)
	fmt.Println("Gas Limit", transferData.Fields[3].Value.(*abi.OptionValue).Value.(*abi.U64Value).Value)
}
