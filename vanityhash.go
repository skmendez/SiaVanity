package siavanity

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	cblake "golang.org/x/crypto/blake2b"
	"io/ioutil"
	"log"
	"lukechampine.com/us/merkle"
	"lukechampine.com/us/merkle/blake2b"
	"math"
	"math/bits"
	"os"
	"runtime"
	"unsafe"
)

///*
//#include <stdio.h>
//#include <stdlib.h>
//#include <stdint.h>
//typedef uint8_t roots[16][32];
//
//
//void pass_struct(roots *in) {
//printf("\n[");
//for (int i=0; i < 32; i++)
//{
//    printf("%d ", (*in)[10][i]);
//}
//printf("\b]\n");
//}
//*/
//import "C"

////#cgo CFLAGS: -I.
////#cgo LDFLAGS: -L. -ltest
////#cgo LDFLAGS: -lcudart
////#include <test.h>
//import "C"

const (
	sectorSize = 1 << 22
	leafSize   = 64
	BlockSize  = 128
	dataSize   = leafSize + 1
	// The hash size of BLAKE2b-512 in bytes.
	Size = 64
)

type Stack struct {
	roots [17][32]byte // ordered smallest-to-largest
	used  uint32       // one bit per Stack elem
}

func (s *Stack) hasNodeAtHeight(height int) bool {
	return s.used&(1<<height) != 0
}

func (s *Stack) appendLeaf(leaf []byte) {
	var buf [64]byte
	copy(buf[:], leaf)
	h := blake2b.SumLeaf(&buf)
	i := 0
	for ; s.hasNodeAtHeight(i); i++ {
		h = blake2b.SumPair(s.roots[i], h)
	}
	s.roots[i] = h
	s.used++
}

func (s *Stack) finalRoot(leaf *[64]byte) [32]byte {
	h := blake2b.SumLeaf(leaf)
	i := 0
	for ; s.hasNodeAtHeight(i); i++ {
		h = blake2b.SumPair(s.roots[i], h)
	}
	return h
}

func (s *Stack) finalRootUnwrapped(leaf *[64]byte) [32]byte {
	h := blake2b.SumLeaf(leaf)
	h = blake2b.SumPair(s.roots[0], h)
	h = blake2b.SumPair(s.roots[1], h)
	h = blake2b.SumPair(s.roots[2], h)
	h = blake2b.SumPair(s.roots[3], h)
	h = blake2b.SumPair(s.roots[4], h)
	h = blake2b.SumPair(s.roots[5], h)
	h = blake2b.SumPair(s.roots[6], h)
	h = blake2b.SumPair(s.roots[7], h)
	h = blake2b.SumPair(s.roots[8], h)
	h = blake2b.SumPair(s.roots[9], h)
	h = blake2b.SumPair(s.roots[10], h)
	h = blake2b.SumPair(s.roots[11], h)
	h = blake2b.SumPair(s.roots[12], h)
	h = blake2b.SumPair(s.roots[13], h)
	h = blake2b.SumPair(s.roots[14], h)
	h = blake2b.SumPair(s.roots[15], h)
	return h
}

func (s *Stack) root() [32]byte {
	i := bits.TrailingZeros32(s.used)
	if i == 32 {
		return [32]byte{}
	}
	root := s.roots[i]
	for i++; i < 32; i++ {
		if s.hasNodeAtHeight(i) {
			root = blake2b.SumPair(s.roots[i], root)
		}
	}
	return root
}

type linkfileLayout struct {
	Version            uint8
	Filesize           uint64
	MetadataSize       uint64
	FanoutSize         uint64
	FanoutDataPieces   uint8
	FanoutParityPieces uint8
	CipherType         [8]byte
	CipherKey          [64]byte
}

type linkfileMetadata struct {
	Filename string
	Mode     os.FileMode
}

func prependHeader(name string, data []byte) []byte {
	metadata, _ := json.Marshal(linkfileMetadata{
		Filename: name,
		Mode:     0666,
	})
	layout := linkfileLayout{
		Version:      1,
		Filesize:     uint64(len(data)),
		MetadataSize: uint64(len(metadata)),
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, layout)
	buf.Write(metadata)
	buf.Write(data)
	return buf.Bytes()
}

//func progress(nonce *uint64, complete *bool) {
//	start := time.Now()
//	for !*complete {
//		time.Sleep(1 * time.Second)
//		fmt.Printf("\rCurrent nonce: %v (%.0f iters/sec)", *nonce, float64(*nonce)/time.Since(start).Seconds())
//	}
//}

//func compare(root [32]byte, prefix []byte) bool {
//	for i := 0; i < len(prefix); i++ {
//		if root[i] != prefix[i] {
//			return false
//		}
//	}
//	return true;
//}

//func compareInt(rootPrefix []byte, prefix uint64, bitlength uint8) bool {
//	rootInt := binary.LittleEndian.Uint64(rootPrefix)
//	var mask uint64 = (1 << bitlength) - 1
//	return (rootInt & mask) == prefix
//}
//
//func compute(stack *Stack, prefixBytes []byte) *[64]byte {
//	var result [64]byte
//	fastrand.Read(result[0:8])
//	if r := stack.finalRootUnwrapped(&result); compare(r, prefixBytes) {
//		return &result
//	}
//	return nil
//}

func FinalLeaf(roots *[16][32]byte, result *[64]byte) [32]byte {
	var buf [128]byte
	copy(buf[1:], result[:])
	h := CheckSum(&buf)
	buf[0] = 1
	for i := 0; i < 16; i++ {
		copy(buf[33:], h[:])
		copy(buf[1:], roots[i][:])
		h = CheckSum(&buf)
	}
	return h
}

func FinalLeafTheirCheckSum(roots *[16][32]byte, result *[64]byte) [32]byte {
	var buf [65]byte
	copy(buf[1:], result[:])
	h := cblake.Sum256(buf[:])
	buf[0] = 1
	for i := 0; i < 16; i++ {
		copy(buf[33:], h[:])
		copy(buf[1:], roots[i][:])
		h = cblake.Sum256(buf[:])
	}
	return h
}

var precomputed = [12][16]byte{
	{0, 2, 4, 6, 1, 3, 5, 7, 8, 10, 12, 14, 9, 11, 13, 15},
	{14, 4, 9, 13, 10, 8, 15, 6, 1, 0, 11, 5, 12, 2, 7, 3},
	{11, 12, 5, 15, 8, 0, 2, 13, 10, 3, 7, 9, 14, 6, 1, 4},
	{7, 3, 13, 11, 9, 1, 12, 14, 2, 5, 4, 15, 6, 10, 0, 8},
	{9, 5, 2, 10, 0, 7, 4, 15, 14, 11, 6, 3, 1, 12, 8, 13},
	{2, 6, 0, 8, 12, 10, 11, 3, 4, 7, 15, 1, 13, 5, 14, 9},
	{12, 1, 14, 4, 5, 15, 13, 10, 0, 6, 9, 8, 7, 3, 2, 11},
	{13, 7, 12, 3, 11, 14, 1, 9, 5, 15, 8, 2, 0, 4, 6, 10},
	{6, 14, 11, 0, 15, 9, 3, 8, 12, 13, 1, 10, 2, 7, 4, 5},
	{10, 8, 7, 1, 2, 4, 6, 5, 15, 9, 3, 13, 11, 14, 12, 0},
	{0, 2, 4, 6, 1, 3, 5, 7, 8, 10, 12, 14, 9, 11, 13, 15}, // equal to the first
	{14, 4, 9, 13, 10, 8, 15, 6, 1, 0, 11, 5, 12, 2, 7, 3}, // equal to the second
}

func CheckSum(blocks *[BlockSize]byte) [32]byte {

	v := [16]uint64{
		0x6a09e667f2bdc928, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
		0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179,
		0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
		0x510e527fade68290, 0x9b05688c2b3e6c1f, 0xe07c265404be4294, 0x5be0cd19137e2179,
	}

	h := [4]uint64{v[0], v[1], v[2], v[3]}

	m := (*[16]uint64)(unsafe.Pointer(blocks))

	for j := range precomputed {
		s := &(precomputed[j])

		v[0] += m[s[0]]
		v[0] += v[4]
		v[12] ^= v[0]
		v[12] = bits.RotateLeft64(v[12], -32)
		v[8] += v[12]
		v[4] ^= v[8]
		v[4] = bits.RotateLeft64(v[4], -24)
		v[0] += m[s[4]]
		v[0] += v[4]
		v[12] ^= v[0]
		v[12] = bits.RotateLeft64(v[12], -16)
		v[8] += v[12]
		v[4] ^= v[8]
		v[4] = bits.RotateLeft64(v[4], -63)

		v[1] += m[s[1]]
		v[1] += v[5]
		v[13] ^= v[1]
		v[13] = bits.RotateLeft64(v[13], -32)
		v[9] += v[13]
		v[5] ^= v[9]
		v[5] = bits.RotateLeft64(v[5], -24)
		v[1] += m[s[5]]
		v[1] += v[5]
		v[13] ^= v[1]
		v[13] = bits.RotateLeft64(v[13], -16)
		v[9] += v[13]
		v[5] ^= v[9]
		v[5] = bits.RotateLeft64(v[5], -63)

		v[2] += m[s[2]]
		v[2] += v[6]
		v[14] ^= v[2]
		v[14] = bits.RotateLeft64(v[14], -32)
		v[10] += v[14]
		v[6] ^= v[10]
		v[6] = bits.RotateLeft64(v[6], -24)
		v[2] += m[s[6]]
		v[2] += v[6]
		v[14] ^= v[2]
		v[14] = bits.RotateLeft64(v[14], -16)
		v[10] += v[14]
		v[6] ^= v[10]
		v[6] = bits.RotateLeft64(v[6], -63)

		v[3] += m[s[3]]
		v[3] += v[7]
		v[15] ^= v[3]
		v[15] = bits.RotateLeft64(v[15], -32)
		v[11] += v[15]
		v[7] ^= v[11]
		v[7] = bits.RotateLeft64(v[7], -24)
		v[3] += m[s[7]]
		v[3] += v[7]
		v[15] ^= v[3]
		v[15] = bits.RotateLeft64(v[15], -16)
		v[11] += v[15]
		v[7] ^= v[11]
		v[7] = bits.RotateLeft64(v[7], -63)

		v[0] += m[s[8]]
		v[0] += v[5]
		v[15] ^= v[0]
		v[15] = bits.RotateLeft64(v[15], -32)
		v[10] += v[15]
		v[5] ^= v[10]
		v[5] = bits.RotateLeft64(v[5], -24)
		v[0] += m[s[12]]
		v[0] += v[5]
		v[15] ^= v[0]
		v[15] = bits.RotateLeft64(v[15], -16)
		v[10] += v[15]
		v[5] ^= v[10]
		v[5] = bits.RotateLeft64(v[5], -63)

		v[1] += m[s[9]]
		v[1] += v[6]
		v[12] ^= v[1]
		v[12] = bits.RotateLeft64(v[12], -32)
		v[11] += v[12]
		v[6] ^= v[11]
		v[6] = bits.RotateLeft64(v[6], -24)
		v[1] += m[s[13]]
		v[1] += v[6]
		v[12] ^= v[1]
		v[12] = bits.RotateLeft64(v[12], -16)
		v[11] += v[12]
		v[6] ^= v[11]
		v[6] = bits.RotateLeft64(v[6], -63)

		v[2] += m[s[10]]
		v[2] += v[7]
		v[13] ^= v[2]
		v[13] = bits.RotateLeft64(v[13], -32)
		v[8] += v[13]
		v[7] ^= v[8]
		v[7] = bits.RotateLeft64(v[7], -24)
		v[2] += m[s[14]]
		v[2] += v[7]
		v[13] ^= v[2]
		v[13] = bits.RotateLeft64(v[13], -16)
		v[8] += v[13]
		v[7] ^= v[8]
		v[7] = bits.RotateLeft64(v[7], -63)

		v[3] += m[s[11]]
		v[3] += v[4]
		v[14] ^= v[3]
		v[14] = bits.RotateLeft64(v[14], -32)
		v[9] += v[14]
		v[4] ^= v[9]
		v[4] = bits.RotateLeft64(v[4], -24)
		v[3] += m[s[15]]
		v[3] += v[4]
		v[14] ^= v[3]
		v[14] = bits.RotateLeft64(v[14], -16)
		v[9] += v[14]
		v[4] ^= v[9]
		v[4] = bits.RotateLeft64(v[4], -63)
	}

	h[0] ^= v[0] ^ v[8]
	h[1] ^= v[1] ^ v[9]
	h[2] ^= v[2] ^ v[10]
	h[3] ^= v[3] ^ v[11]

	return *(*[32]byte)(unsafe.Pointer(&h))
}

func round(v0, v4, v8, v12 *uint64, m *[16]uint64, s *[16]byte, c int) {
	*v0 += m[s[c]]
	*v0 += *v4
	*v12 ^= *v0
	*v12 = bits.RotateLeft64(*v12, -32)
	*v8 += *v12
	*v4 ^= *v8
	*v4 = bits.RotateLeft64(*v4, -24)
	*v0 += m[s[c+4]]
	*v0 += *v4
	*v12 ^= *v0
	*v12 = bits.RotateLeft64(*v12, -16)
	*v8 += *v12
	*v4 ^= *v8
	*v4 = bits.RotateLeft64(*v4, -63)

}

func round1(va, vb, vc, vd *uint64, m *[16]uint64, s *[16]byte, c int) {
	*va += m[s[c]]
	*va += *vb
	*vd ^= *va
	*vd = bits.RotateLeft64(*vd, -32)
	*vc += *vd
	*vb ^= *vc
	*vb = bits.RotateLeft64(*vb, -24)
}

func round2(va, vb, vc, vd *uint64, m *[16]uint64, s *[16]byte, c int) {
	*va += m[s[c+4]]
	*va += *vb
	*vd ^= *va
	*vd = bits.RotateLeft64(*vd, -16)
	*vc += *vd
	*vb ^= *vc
	*vb = bits.RotateLeft64(*vb, -63)
}

func CheckSum2(blocks *[BlockSize]byte) [32]byte {

	v := [16]uint64{
		0x6a09e667f2bdc928, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
		0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179,
		0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
		0x510e527fade68290, 0x9b05688c2b3e6c1f, 0xe07c265404be4294, 0x5be0cd19137e2179,
	}

	h := [4]uint64{v[0], v[1], v[2], v[3]}

	m := (*[16]uint64)(unsafe.Pointer(blocks))

	for j := range precomputed {
		s := &(precomputed[j])

		round(&v[0], &v[4], &v[8], &v[12], m, s, 0)
		round(&v[1], &v[5], &v[9], &v[13], m, s, 1)
		round(&v[2], &v[6], &v[10], &v[14], m, s, 2)
		round(&v[3], &v[7], &v[11], &v[15], m, s, 3)

		round(&v[0], &v[5], &v[10], &v[15], m, s, 8)

		round(&v[1], &v[6], &v[11], &v[12], m, s, 9)
		round(&v[2], &v[7], &v[8], &v[13], m, s, 10)
		round(&v[3], &v[4], &v[9], &v[14], m, s, 11)

		//round1(&v[0], &v[4], &v[8], &v[12], m, s, 0)
		//round2(&v[0], &v[4], &v[8], &v[12], m, s, 0)
		//round1(&v[1], &v[5], &v[9], &v[13], m, s, 1)
		//round2(&v[1], &v[5], &v[9], &v[13], m, s, 1)
		//round1(&v[2], &v[6], &v[10], &v[14], m, s, 2)
		//round2(&v[2], &v[6], &v[10], &v[14], m, s, 2)
		//round1(&v[3], &v[7], &v[11], &v[15], m, s, 3)
		//round2(&v[3], &v[7], &v[11], &v[15], m, s, 3)
		//round1(&v[0], &v[5], &v[10], &v[15], m, s, 8)
		//round2(&v[0], &v[5], &v[10], &v[15], m, s, 8)
		//round1(&v[1], &v[6], &v[11], &v[12], m, s, 9)
		//round2(&v[1], &v[6], &v[11], &v[12], m, s, 9)
		//round1(&v[2], &v[7], &v[8], &v[13], m, s, 10)
		//round2(&v[2], &v[7], &v[8], &v[13], m, s, 10)
		//round1(&v[3], &v[4], &v[9], &v[14], m, s, 11)
		//round2(&v[3], &v[4], &v[9], &v[14], m, s, 11)
	}

	h[0] ^= v[0] ^ v[8]
	h[1] ^= v[1] ^ v[9]
	h[2] ^= v[2] ^ v[10]
	h[3] ^= v[3] ^ v[11]

	return *(*[32]byte)(unsafe.Pointer(&h))
}

func CheckSum3(blocks *[BlockSize]byte) [32]byte {

	v := [16]uint64{
		0x6a09e667f2bdc928, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
		0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179,
		0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
		0x510e527fade68290, 0x9b05688c2b3e6c1f, 0xe07c265404be4294, 0x5be0cd19137e2179,
	}

	h := [4]uint64{v[0], v[1], v[2], v[3]}

	m := (*[16]uint64)(unsafe.Pointer(blocks))

	for j := range precomputed {
		s := &(precomputed[j])

		//round(&v[0], &v[4], &v[8], &v[12], m, s, 0)
		//round(&v[1], &v[5], &v[9], &v[13], m, s, 1)
		//round(&v[2], &v[6], &v[10], &v[14], m, s, 2)
		//round(&v[3], &v[7], &v[11], &v[15], m, s, 3)
		//
		//round(&v[0], &v[5], &v[10], &v[15], m, s, 8)
		//
		//round(&v[1], &v[6], &v[11], &v[12], m, s, 9)
		//round(&v[2], &v[7], &v[8], &v[13], m, s, 10)
		//round(&v[3], &v[4], &v[9], &v[14], m, s, 11)

		round1(&v[0], &v[4], &v[8], &v[12], m, s, 0)
		round2(&v[0], &v[4], &v[8], &v[12], m, s, 0)
		round1(&v[1], &v[5], &v[9], &v[13], m, s, 1)
		round2(&v[1], &v[5], &v[9], &v[13], m, s, 1)
		round1(&v[2], &v[6], &v[10], &v[14], m, s, 2)
		round2(&v[2], &v[6], &v[10], &v[14], m, s, 2)
		round1(&v[3], &v[7], &v[11], &v[15], m, s, 3)
		round2(&v[3], &v[7], &v[11], &v[15], m, s, 3)
		round1(&v[0], &v[5], &v[10], &v[15], m, s, 8)
		round2(&v[0], &v[5], &v[10], &v[15], m, s, 8)
		round1(&v[1], &v[6], &v[11], &v[12], m, s, 9)
		round2(&v[1], &v[6], &v[11], &v[12], m, s, 9)
		round1(&v[2], &v[7], &v[8], &v[13], m, s, 10)
		round2(&v[2], &v[7], &v[8], &v[13], m, s, 10)
		round1(&v[3], &v[4], &v[9], &v[14], m, s, 11)
		round2(&v[3], &v[4], &v[9], &v[14], m, s, 11)
	}

	h[0] ^= v[0] ^ v[8]
	h[1] ^= v[1] ^ v[9]
	h[2] ^= v[2] ^ v[10]
	h[3] ^= v[3] ^ v[11]

	return *(*[32]byte)(unsafe.Pointer(&h))
}

//func main1() {
//	fmt.Printf("Invoking cuda library...\n")
//	//fmt.Println("Done ", C.test_add())
//}
//
//
//func main2() {
//	leaf := [64]byte{}
//	fastrand.Read(leaf[:])
//	fmt.Println(blake2b.SumLeaf(&leaf))
//	var buf [128]byte
//	buf[0] = byte(0)
//	copy(buf[1:], leaf[:])
//	sum := CheckSum(&buf)
//	fmt.Println(sum[:32])
//}
//
//func main() {
//
//
//}
//
//
//
//
//
//
func main() {

	fileData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fileData = prependHeader(os.Args[1], fileData)
	if len(fileData) > sectorSize-leafSize {
		log.Fatal("file is too large")
	}
	// construct intermediate Stack
	var base Stack
	b := bytes.NewBuffer(fileData)
	for base.used != (sectorSize/leafSize)-1 {
		base.appendLeaf(b.Next(leafSize))
	}
	prefix := os.Args[2]
	prefixLen := len(prefix)
	difficulty := math.Pow(64, float64(prefixLen))
	var roots [16][32]byte
	copy(roots[:], base.roots[:])
	//C.pass_struct((*C.roots)(unsafe.Pointer(&roots)))
	fmt.Printf("Expected difficulty: %.0f\n", difficulty)
	finder := MultiThread{runtime.NumCPU()}
	leaf := finder.find(roots, prefix)
	// double-check
	var sector [sectorSize]byte
	copy(sector[:], fileData)
	copy(sector[len(sector)-leafSize:], leaf[:])
	root := merkle.SectorRoot(&sector)

	fmt.Println("\nFinished!")
	fmt.Printf("Produces Merkle root %s", base64.RawURLEncoding.EncodeToString(root[:]))

	dst := os.Args[1] + ".rawsector"
	if err := ioutil.WriteFile(dst, sector[:], 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nWrote raw sector to", dst)
}
