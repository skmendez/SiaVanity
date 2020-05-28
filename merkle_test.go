package siavanity

import (
	"bytes"
	"gitlab.com/NebulousLabs/fastrand"
	"testing"
	"unsafe"
)

func setup(fileSize int) ([16][32]byte, [64]byte, Stack) {
	fileData := fastrand.Bytes(fileSize)
	var base Stack
	b := bytes.NewBuffer(fileData)
	for base.used != (sectorSize/leafSize)-1 {
		base.appendLeaf(b.Next(leafSize))
	}
	var roots [16][32]byte
	copy(roots[:], base.roots[:])
	fastrand.Read((*[512]byte)(unsafe.Pointer(&roots))[:])
	var result [64]byte
	fastrand.Read(result[:])
	return roots, result, base
}

//var re [32]byte

func BenchmarkFinalLeaf(b *testing.B) {
	roots, result, _ := setup(1000)
	var r [32]byte
	for n := 0; n < b.N; n++ {
		r = FinalLeaf(&roots, &result)
	}
	re = r
}

func BenchmarkFinalLeafTheirVersion(b *testing.B) {
	roots, result, _ := setup(1000)
	var r [32]byte
	for n := 0; n < b.N; n++ {
		r = FinalLeafTheirCheckSum(&roots, &result)
	}
	re = r
}

func BenchmarkMyOptimized(b *testing.B) {
	_, result, stack := setup(1000)
	var r [32]byte
	for n := 0; n < b.N; n++ {
		r = stack.finalRoot(&result)
	}
	re = r
}

func BenchmarkOriginal(b *testing.B) {
	_, result, stack := setup(1000)
	var r [32]byte
	for n := 0; n < b.N; n++ {
		s := stack
		s.appendLeaf(result[:])
		r = s.root()
	}
	re = r
}
