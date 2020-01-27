#!/bin/bash
go test -bench="Benchmark_AddWithEviction_UniformDistribution_250000x1_WithConfig_NumSendersToEvictInOneStep_10$" -benchtime 1x
go test -bench="Benchmark_AddWithEviction_UniformDistribution_250000x1_WithConfig_NumSendersToEvictInOneStep_100$" -benchtime 1x
go test -bench="Benchmark_AddWithEviction_UniformDistribution_250000x1_WithConfig_NumSendersToEvictInOneStep_1000$" -benchtime 1x
go test -bench="Benchmark_AddWithEviction_UniformDistribution_10x25000$" -benchtime 1x
go test -bench="Benchmark_AddWithEviction_UniformDistribution_1x250000$" -benchtime 1x
go test -bench="Benchmark_GetSnapshotAscending$" -benchtime=1x