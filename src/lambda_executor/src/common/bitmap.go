package common

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
	b.Data[pos] |= 1 << off
}

func (b *Bitmap) Get(v int) bool {
	pos, off := v/32, v%32
	return b.Data[pos]&(1<<off) > 0
}
