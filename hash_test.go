package siavanity

import (
	"fmt"
	"gitlab.com/NebulousLabs/fastrand"
	"golang.org/x/crypto/blake2b"
	"testing"
)

func setup2() (out [65]byte) {
	fastrand.Read(out[:])
	return
}

func arrayEquals(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCheckSum(t *testing.T) {
	im := setup2()
	var result [128]byte
	copy(im[:], result[:])
	expected := blake2b.Sum256(im[:])
	actual := CheckSum(&result)
	if !arrayEquals(actual[:], expected[:]) {
		fmt.Println(expected)
		t.Fail()
	}
}

func TestCheckSum2(t *testing.T) {
	for x := 0; x < 1000; x++ {
		im := setup2()
		var result [128]byte
		copy(im[:], result[:])
		expected := blake2b.Sum256(im[:])
		actual := CheckSum2(&result)
		if !arrayEquals(actual[:], expected[:]) {
			t.Fatalf("Failed, expected:\n%v\nbut got:\n%v\n", expected, actual)
		}
	}
}

var re [32]byte

func BenchmarkCheckSum(b *testing.B) {
	im := setup2()
	var result [128]byte
	copy(im[:], result[:])
	var r [32]byte
	for i := 0; i < b.N; i++ {
		r = CheckSum(&result)
	}
	re = r
}

func BenchmarkCheckSum2(b *testing.B) {
	im := setup2()
	var result [128]byte
	copy(im[:], result[:])
	var r [32]byte
	for i := 0; i < b.N; i++ {
		r = CheckSum2(&result)
	}
	re = r
}

func BenchmarkCheckSum3(b *testing.B) {
	im := setup2()
	var result [128]byte
	copy(im[:], result[:])
	var r [32]byte
	for i := 0; i < b.N; i++ {
		r = CheckSum3(&result)
	}
	re = r
}

func BenchmarkAUX2CheckSum(b *testing.B) {
	im := setup2()
	var r [32]byte
	for i := 0; i < b.N; i++ {
		r = blake2b.Sum256(im[:])
	}
	re = r
}
