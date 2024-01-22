package main

import (
	"reflect"
	"time"
	"unsafe"
)

func GetTime() float64 {
	return float64(time.Now().UnixNano()) / 1e9
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
