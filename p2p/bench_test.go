package main

import (
	"testing"
	"time"
)

func Benchmark01(t *testing.B) {

	t.Run("bench milli", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Sleep(time.Millisecond)
		}
	})

	t.Run("bench micro", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Sleep(time.Microsecond)
		}
	})

	t.Run("bench nano", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Sleep(time.Nanosecond)
		}
	})
}

type A struct {
	field1 int "this is field1"
	field2 int
}
