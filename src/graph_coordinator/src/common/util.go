package common

import (
	"reflect"
	"time"
	"unsafe"
)

func GetTime() float64 {
	return float64(time.Now().UnixNano()) / 1e9
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
