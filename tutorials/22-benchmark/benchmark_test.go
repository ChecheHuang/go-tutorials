package main

import "testing"

func BenchmarkConcatPlus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConcatPlus(1000)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConcatBuilder(1000)
	}
}

func BenchmarkSliceAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SliceAppend(10000)
	}
}

func BenchmarkSlicePrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SlicePrealloc(10000)
	}
}
