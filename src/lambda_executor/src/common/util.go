package common

import (
	"encoding/binary"
	"reflect"
	"time"
	"unsafe"
)

func Uint32ToBytes(n uint32, b []byte) {
	binary.LittleEndian.PutUint32(b, n)
}

func BytesToUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func Int32ToBytes(n int32, b []byte) {
	binary.LittleEndian.PutUint32(b, uint32(n))
}

func BytesToInt32(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b))
}

func Uint32Slice2ByteSlice(a []uint32) []byte {
	var b []byte
	uint32Slice := reflect.ValueOf(a)
	byteSlice := reflect.New(reflect.TypeOf(b))
	header := (*reflect.SliceHeader)(unsafe.Pointer(byteSlice.Pointer()))
	header.Cap = uint32Slice.Cap() * 4
	header.Len = uint32Slice.Len() * 4
	header.Data = uintptr(uint32Slice.Pointer())
	return byteSlice.Elem().Bytes()
}

func ByteSlice2Uint32Slice(a []byte) []uint32 {
	var u []uint32
	byteSlice := reflect.ValueOf(a)
	uint32Slice := reflect.New(reflect.TypeOf(u))
	header := (*reflect.SliceHeader)(unsafe.Pointer(uint32Slice.Pointer()))
	header.Cap = byteSlice.Cap() / 4
	header.Len = byteSlice.Len() / 4
	header.Data = uintptr(byteSlice.Pointer())
	return uint32Slice.Elem().Interface().([]uint32)
}

func ByteSlice2Int32Slice(a []byte) []int32 {
	var u []int32
	byteSlice := reflect.ValueOf(a)
	int32Slice := reflect.New(reflect.TypeOf(u))
	header := (*reflect.SliceHeader)(unsafe.Pointer(int32Slice.Pointer()))
	header.Cap = byteSlice.Cap() / 4
	header.Len = byteSlice.Len() / 4
	header.Data = uintptr(byteSlice.Pointer())
	return int32Slice.Elem().Interface().([]int32)
}

func ByteSlice2Int64Slice(a []byte) []int64 {
	var u []int64
	byteSlice := reflect.ValueOf(a)
	int64Slice := reflect.New(reflect.TypeOf(u))
	header := (*reflect.SliceHeader)(unsafe.Pointer(int64Slice.Pointer()))
	header.Cap = byteSlice.Cap() / 8
	header.Len = byteSlice.Len() / 8
	header.Data = uintptr(byteSlice.Pointer())
	return int64Slice.Elem().Interface().([]int64)
}

func GetTime() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}
