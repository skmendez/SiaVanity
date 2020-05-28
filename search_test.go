package siavanity

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

var roots = [16][32]byte{}

var prefix = "eeeee"

var result [64]byte

func BenchmarkSearch(b *testing.B) {
	var r [64]byte
	finders := map[string]VanityFinder{
		"SingleThread": SingleThread{},
		"4Thread":      MultiThread{4},
		"12Thread":     MultiThread{12},
		"36Thread":     MultiThread{36},
		"4Atomic":      MultiThreadAtomic{4},
		"12Atomic":     MultiThreadAtomic{12},
		"36Atomic":     MultiThreadAtomic{36},
		"4Progress":    MultiThreadProgress{4},
		"12Progress":   MultiThreadProgress{12},
		"36Progress":   MultiThreadProgress{36},
	}

	for name, bm := range finders {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r = bm.find(roots, prefix)
			}
			result = r
		})
	}
}

func BenchmarkBasicSearch(b *testing.B) {
	//var r [64]byte
	//for n := 0; n < b.N; n++ {
	//	r = BasicSearch(roots, "eeeee")
	//}
	//result = r
}

func BenchmarkCountToOnePointTwoBillion(b *testing.B) {
	const ONEB = 1_200_000_000
	const THREAD_COUNT = 120

	var a uint64
	for n := 0; n < b.N; n++ {
		atomic.StoreUint64(&a, 0)
		wg := sync.WaitGroup{}
		for i := 0; i < THREAD_COUNT; i++ {
			wg.Add(1)
			go func() {
				for j := 0; j < ONEB/THREAD_COUNT; j++ {
					atomic.AddUint64(&a, 1)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		if atomic.LoadUint64(&a) != ONEB {
			panic(a)
		}
	}
}

func TestProgress(t *testing.T) {
	search := MultiThreadProgress{12}
	fmt.Println(search.find(roots, "eeee"))
}
