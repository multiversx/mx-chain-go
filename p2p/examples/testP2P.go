package main

import "github.com/ElrondNetwork/elrond-go-sandbox/p2p/ex09BenchmarkDirectSendMultiple"

//import "elrond-go-sandbox/p2p/ex09BenchmarkDirectSendMultiple"

//a import "./ex01ChatLibP2P"
//a import "./ex02AdvChatLibP2P"
//a import "./ex03ONetTesting"
//a import "./ex04TestDiscovery"
//a import "./ex05ChatWithGossipPeerDiscovery"
//import "./ex06Floodsub"

func main() {
	//comment and uncomment which test do you want to run

	//example 1, simple one-way communication with limited benchmarking
	//ex01ChatLibP2P.Main();

	//example 2, advanced multi node chat with benchmarking
	//ex02AdvChatLibP2P.Main();

	//example 3, test onet
	//ex03ONetTesting.Main()

	//example 4, test network disovery on libP2P
	//ex04TestDiscovery.Main()

	//example 5, test peer discovery via gossip protocol
	//ex05ChatWithGossipPeerDiscovery.Main()

	//example 6, test floodsub implementation
	//ex06Floodsub.Main()

	//example 7, test peer chat and gossip
	//ex07ChatWithGossip.Main()

	//example 8, benchmark libP2P against multiple nodes
	//ex08BenchmarkSendPackets.Main()

	ex09BenchmarkDirectSendMultiple.Main()
}
