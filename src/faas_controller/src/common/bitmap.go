package common

import (
	"fmt"
)

//go:noescape
func Or(ptr *uint32, val uint32)

type Bitmap struct {
	Data []uint32
}

func NewBitmap(size int) *Bitmap {
	return &Bitmap{Data: make([]uint32, size/32+1)}
}

func NewBitmapWith(data []uint32) *Bitmap {
	return &Bitmap{Data: data}
}

func (b *Bitmap) Add(v int) {
	pos, off := v/32, v%32
	Or(&b.Data[pos], 1<<off)
}

func (b *Bitmap) Get(v int) bool {
	pos, off := v/32, v%32
	return b.Data[pos]&(1<<off) > 0
}

func (b *Bitmap) Or(c *Bitmap) {
	for i := 0; i < len(b.Data); i++ {
		Or(&b.Data[i], c.Data[i])
	}
}

func (b *Bitmap) Or2(c *Bitmap) {
	for i := 0; i < len(b.Data); i++ {
		b.Data[i] |= c.Data[i]
	}
}

func (b *Bitmap) And(c *Bitmap) {
	for i := 0; i < len(b.Data); i++ {
		b.Data[i] &= c.Data[i]
	}
}

func (b *Bitmap) Empty() bool {
	for i := 0; i < len(b.Data); i++ {
		if b.Data[i] > 0 {
			return false
		}
	}
	return true
}

func (b *Bitmap) Print() {
	for i := 0; i < len(b.Data); i++ {
		fmt.Printf("%b ", b.Data[i])
	}
	fmt.Println()
}

func (b *Bitmap) Count() int {
	var cnt int
	for _, d := range b.Data {
		v := int(d)
		for v != 0 {
			cnt += v & 1
			v >>= 1
		}
	}
	return cnt
}
