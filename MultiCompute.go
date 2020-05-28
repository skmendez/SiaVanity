package siavanity

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type VanityFinder interface {
	find(roots [16][32]byte, prefix string) [64]byte
}

func searchThread(numThreads int, threadId int, roots [16][32]byte, value chan [64]byte, signal chan struct{}, prefix string) {
	nonce := uint64(threadId)
	var result [64]byte
loop:
	for {
		select {
		case <-signal:
			//fmt.Printf("channel closed, current nonce: %v\n", nonce)
			break loop
		default:
			binary.LittleEndian.PutUint64(result[:], nonce)
			checkSum := FinalLeafTheirCheckSum(&roots, &result)
			if compare(checkSum, prefix) {
				//fmt.Printf("Winner!: Thread %v with nonce %v\n", threadId, nonce)
				value <- result
				break loop
			}
			nonce += uint64(numThreads)
		}
	}
}

type MultiThread struct {
	numThreads int
}

func (m MultiThread) find(roots [16][32]byte, prefix string) [64]byte {
	signal := make(chan struct{})
	value := make(chan [64]byte)
	for t := 0; t < m.numThreads; t++ {
		go searchThread(m.numThreads, t, roots, value, signal, prefix)
	}
	result := <-value
	close(signal)
	return result
}

type searchWorker struct {
	numThreads   int
	threadId     int
	signal       chan struct{}
	progressSend chan uint64
}

func (s searchWorker) search(roots [16][32]byte, prefix string, value chan [64]byte) {
	var hashCount uint64
	var result [64]byte
loop:
	for {
		select {
		case _, ok := <-s.signal:
			if !ok {
				return
			} else {
				s.progressSend <- hashCount
			}
		default:
			binary.LittleEndian.PutUint64(result[:], uint64(s.threadId)+hashCount*uint64(s.numThreads))
			checkSum := FinalLeafTheirCheckSum(&roots, &result)
			if compare(checkSum, prefix) {
				value <- result
				close(s.progressSend)
				break loop
			}
			hashCount += 1
		}
	}
}

func gatherProgress(workers []searchWorker) (hashCount uint64, err error) {
	for _, worker := range workers {
		worker.signal <- struct{}{}
	}

	for _, worker := range workers {
		count, ok := <-worker.progressSend
		if !ok {
			return 0, errors.New("worker closed")
		}
		hashCount += count
	}
	return
}

type MultiThreadProgress struct {
	numThreads int
}

func threadProgress(signalComplete chan struct{}, workers []searchWorker, done chan struct{}) {
	defer close(done)
	start := time.Now()
	for {
		select {
		case <-signalComplete:
			for _, worker := range workers {
				close(worker.signal)
			}
			return
		default:
			time.Sleep(time.Second)
			progress, err := gatherProgress(workers)
			if err == nil {
				fmt.Sprintf("%v\r", float64(progress)/time.Since(start).Seconds())
			}
		}
	}
}

func (m MultiThreadProgress) find(roots [16][32]byte, prefix string) [64]byte {
	workers := make([]searchWorker, m.numThreads)
	value := make(chan [64]byte)
	for i := range workers {
		workers[i].threadId = i
		workers[i].numThreads = m.numThreads
		workers[i].progressSend = make(chan uint64)
		workers[i].signal = make(chan struct{}, 2)
		go workers[i].search(roots, prefix, value)
	}

	signalComplete := make(chan struct{})
	done := make(chan struct{})
	go threadProgress(signalComplete, workers, done)
	result := <-value
	close(signalComplete)
	<-done
	return result
}

type MultiThreadAtomic struct {
	numThreads int
}

func (m MultiThreadAtomic) find(roots [16][32]byte, prefix string) [64]byte {
	var nonce uint64
	var value = make(chan [64]byte)
	signal := make(chan struct{})
	for t := 0; t < m.numThreads; t++ {
		go searchThreadAtomic(&nonce, t, roots, value, signal, prefix)
	}
	result := <-value
	close(signal)
	return result
}

func searchThreadAtomic(nonceRef *uint64, threadId int, roots [16][32]byte, value chan [64]byte, signal chan struct{}, prefix string) {
	var result [64]byte
loop:
	for {
		var mynonce uint64
		select {
		case <-signal:
			//fmt.Printf("channel closed, current nonce: %v\n", mynonce)
			break loop
		default:
			mynonce = atomic.AddUint64(nonceRef, 1)
			binary.LittleEndian.PutUint64(result[:], mynonce)
			checkSum := FinalLeafTheirCheckSum(&roots, &result)
			if compare(checkSum, prefix) {
				//fmt.Printf("Winner!: Thread %v with nonce %v\n", threadId, mynonce)
				value <- result
				break loop
			}
		}
	}
}

func compare(checkSum [32]byte, prefix string) bool {
	return strings.HasPrefix(base64.RawURLEncoding.EncodeToString(checkSum[:]), prefix)
}

type SingleThread struct{}

func (s SingleThread) find(roots [16][32]byte, prefix string) [64]byte {
	nonce := uint64(0)
	var result [64]byte
	for {
		binary.LittleEndian.PutUint64(result[:], nonce)
		nonce += 1
		checkSum := FinalLeafTheirCheckSum(&roots, &result)
		if compare(checkSum, prefix) {
			return result
		}
	}
}
