package wire

import (
	"encoding/binary"

	crunch "github.com/superwhiskers/crunch/v3"
)

func ReadIV64Next(b *crunch.Buffer) int64 {
	buf := make([]byte, 0)
	for i := 0; i < binary.MaxVarintLen64; i++ {
		buf = append(buf, b.ReadByteNext())
		v, n := binary.Varint(buf)
		if n == 0 { //Buffer too small
			continue
		}
		if n < 0 {
			panic("readIVarNext: overflow 64 bits")
		}
		return v
	}
	panic("readIVarNext: failed to find end of i64")
}
func WriteIV64Next(b *crunch.Buffer, x int64) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, x)
	b.Grow(int64(n))
	b.WriteBytesNext(buf[:n])
}

func ReadIV64(p []byte) int64 {
	v, n := binary.Varint(p)
	if n == 0 {
		panic("readIVar: buffer too small")
	}
	if n < 0 {
		panic("readIVar: overflow 64 bits")
	}
	return v
}

func WriteIV64(x int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, x)
	return buf[:n]
}
