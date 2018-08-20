package main

import (
	"BenchmarkBroadcast/broadcast"

	"fmt"
)

func main() {
	var nodes, latency, peersPerNode, bandwidth, blocksize int

	fmt.Println("How many peers do you have in the system?")
	fmt.Scan(&nodes)

	fmt.Println("How much time does it take to send a block? (in ms)")
	fmt.Scan(&latency)

	fmt.Println("How many peers does a node need to remember?")
	fmt.Scan(&peersPerNode)

	fmt.Println("What is the peer bandwidth? (in kb)")
	fmt.Scan(&bandwidth)

	fmt.Println("What is the size of a block? (in kb)")
	fmt.Scan(&blocksize)

	latency, nrOfHops := broadcast.Broadcast(nodes, latency, peersPerNode, bandwidth, blocksize)

	fmt.Printf("Nr of hops %v and system latency %v ms\n", nrOfHops, latency)

}
