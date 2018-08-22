package main

import (
	"BenchmarkBroadcast/broadcast"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"strconv"
	"time"
)

func main() {
	RunOnce()
	//RunBenchmark()

}
func RunOnce() {
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

	totalLatency, nrOfHops, lonelyNodes := broadcast.Broadcast(nodes, latency, peersPerNode, bandwidth, blocksize)

	fmt.Println("Total latency", totalLatency, "number of hops", nrOfHops, "lonelyNodes", lonelyNodes)
}

func RunBenchmark() {

	xlsx := excelize.NewFile()
	index := xlsx.NewSheet("Sheet1")
	xlsx.SetCellValue("Sheet1", "A1", "Number of nodes")
	xlsx.SetCellValue("Sheet1", "B1", "Latency")
	xlsx.SetCellValue("Sheet1", "C1", "Peers / node")
	xlsx.SetCellValue("Sheet1", "D1", "Bandwidth (kB/s)")
	xlsx.SetCellValue("Sheet1", "E1", "Block size (kB)")

	xlsx.SetCellValue("Sheet1", "F1", "Number of hops ")
	xlsx.SetCellValue("Sheet1", "G1", "Total latency (ms)")
	xlsx.SetCellValue("Sheet1", "H1", "Number of lonely nodes")

	row := 2
	ct := 10.0

	for nodes := 200; nodes <= 600; nodes += 100 {
		for latency := 100; latency <= 500; latency += 50 {
			for peersPerNode := 4; peersPerNode <= 20; peersPerNode++ {
				for bandwidth := 1024; bandwidth <= 8192; bandwidth += 1024 {
					for blocksize := 128; blocksize <= 1024; blocksize += 128 {

						meanTotalLatency, meanNrOfHops, meanLonelyNodes := 0.0, 0.0, 0.0
						for i := 0; i < 10; i++ {
							totalLatency, nrOfHops, lonelyNodes := broadcast.Broadcast(nodes, latency, peersPerNode, bandwidth, blocksize)
							meanTotalLatency += (float64)(totalLatency)
							meanNrOfHops += (float64)(nrOfHops)
							meanLonelyNodes += (float64)(lonelyNodes)

						}

						xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(row), nodes)
						xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(row), latency)
						xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(row), peersPerNode)
						xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(row), bandwidth)
						xlsx.SetCellValue("Sheet1", "E"+strconv.Itoa(row), blocksize)

						xlsx.SetCellValue("Sheet1", "F"+strconv.Itoa(row), meanNrOfHops/ct)
						xlsx.SetCellValue("Sheet1", "G"+strconv.Itoa(row), meanTotalLatency/ct)
						xlsx.SetCellValue("Sheet1", "H"+strconv.Itoa(row), meanLonelyNodes/ct)

						row++

						fmt.Println(row)

					}
				}

			}

		}
	}

	xlsx.SetActiveSheet(index)

	err := xlsx.SaveAs("./" + time.Now().Format("2006-01-02 15:04:05 ") + "benchmark.xlsx")
	if err != nil {
		fmt.Println(err)
	}
}
