package hash

import (
	"math/bits"
	"unsafe"
)

const BlockSize = 128

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
