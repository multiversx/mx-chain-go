package main

import (
	"BroadcastBenchmark/broadcast"
	//"BroadcastBenchmark/test"
	"fmt"
)

func main() {
	fmt.Println("Hi!")
	//test.Test(7)
	broadcast.Broadcast(10, 3, 4, 2, 1)

}
