package main

import (
	"fmt"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ElrondNetwork/elrond-go-sandbox/benchmark-broadcast/broadcast"
)

func main() {

	totalLatency, nrOfHops, lonelyNodes := broadcast.Broadcast(500, 0.1, 12, 2048, 64)

	fmt.Println("latency", totalLatency, " hops ", nrOfHops, " lonely ", lonelyNodes)
}

func intraShard() {
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet("Sheet1")
	xlsx.SetCellValue("Sheet1", "A1", "Number of nodes")
	xlsx.SetCellValue("Sheet1", "B1", "Latency (s)")
	xlsx.SetCellValue("Sheet1", "C1", "Peers / node")
	xlsx.SetCellValue("Sheet1", "D1", "Bandwidth (kB/s)")
	xlsx.SetCellValue("Sheet1", "E1", "Block size (kB)")

	xlsx.SetCellValue("Sheet1", "F1", "Number of hops ")
	xlsx.SetCellValue("Sheet1", "G1", "Total latency (s)")
	xlsx.SetCellValue("Sheet1", "H1", "Number of lonely nodes")

	row := 2
	ct := 10.0

	for nodes := 500; nodes <= 600; nodes += 100 {
		for latency := 0.1; latency <= 0.2; latency += 0.1 {
			for peersPerNode := 4; peersPerNode <= 20; peersPerNode++ {
				for bandwidth := 2048; bandwidth <= 4096; bandwidth += 1024 {
					for blocksize := 256; blocksize <= 512; blocksize += 128 {

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
			xlsx.SetActiveSheet(index)

			err := xlsx.SaveAs("./benchmark.xlsx")
			if err != nil {
				fmt.Println(err)
			}

		}
	}
}

func interShard() {
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet("Sheet1")
	xlsx.SetCellValue("Sheet1", "A1", "Number of nodes")
	xlsx.SetCellValue("Sheet1", "B1", "Latency (s)")
	xlsx.SetCellValue("Sheet1", "C1", "Peers / node")
	xlsx.SetCellValue("Sheet1", "D1", "Bandwidth (kB/s)")
	xlsx.SetCellValue("Sheet1", "E1", "Block size (kB)")

	xlsx.SetCellValue("Sheet1", "F1", "Number of hops ")
	xlsx.SetCellValue("Sheet1", "G1", "Total latency (s)")
	xlsx.SetCellValue("Sheet1", "H1", "Number of lonely nodes")

	latency := 0.2
	peersPerNode := 14
	bandwidth := 2048
	blocksize := 256
	row := 2
	for nodes := 15000; nodes <= 20000; nodes += 5000 {
		for i := 0; i < 10; i++ {
			totalLatency, nrOfHops, lonelyNodes := broadcast.Broadcast(nodes, latency, peersPerNode, bandwidth, blocksize)
			xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(row), nodes)
			xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(row), latency)
			xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(row), peersPerNode)
			xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(row), bandwidth)
			xlsx.SetCellValue("Sheet1", "E"+strconv.Itoa(row), blocksize)

			xlsx.SetCellValue("Sheet1", "F"+strconv.Itoa(row), nrOfHops)
			xlsx.SetCellValue("Sheet1", "G"+strconv.Itoa(row), totalLatency)
			xlsx.SetCellValue("Sheet1", "H"+strconv.Itoa(row), lonelyNodes)

			xlsx.SetActiveSheet(index)

			err := xlsx.SaveAs("./benchmark1.xlsx")
			if err != nil {
				fmt.Println(err)
			}
			row++
			fmt.Println(row)
		}
	}
}
