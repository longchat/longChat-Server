package util

import (
	"unsafe"
)

func Int2Bytes(i int64) []byte {
	b := (*[8]byte)(unsafe.Pointer(&i))
	return (*b)[:]
}

func Bytes2Int(b []byte) int64 {
	if len(b) != 8 {
		return 0
	}
	i := (*int64)(unsafe.Pointer(&b[0]))
	return *i
}
