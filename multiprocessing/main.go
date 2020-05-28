package main

import (
	"fmt"
	"time"
)

func randUntilZero(end chan bool, signal chan bool, difficulty uint64) {
	for {
		select {
		case <-end:
			break
		default:
			var a uint64
			for a < difficulty {
				a += 1
			}
			signal <- true
			break
		}
	}
}

func main() {
	maxroutines := 16
	runcount := 20
	difficulty := 4e8

	for count := 1; count <= maxroutines; count += 2 {
		start := time.Now()
		for run := 0; run < runcount; run++ {
			SpeedTest(count, uint64(difficulty))
		}
		fmt.Printf("For %v routines it takes %v\n", count, time.Since(start))
	}
}

func SpeedTest(numroutines int, difficulty uint64) {
	end := make(chan bool)
	signal := make(chan bool)
	for i := 0; i < numroutines; i++ {
		go randUntilZero(end, signal, difficulty/uint64(numroutines))
	}
	<-signal
	//for i := 0; i < numroutines; i++ {
	//	end <- complete
	//}
	//fmt.Println("Complete!")
}
