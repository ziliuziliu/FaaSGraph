package graphutil

import (
	"math"
	"strconv"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	IPC_CREATE = 00001000
)

func shmGet(key uintptr, length uintptr) uintptr {
	id, _, err := unix.Syscall(unix.SYS_SHMGET, key, length, IPC_CREATE|0666)
	if err != 0 {
		panic(err)
	}
	return id
}

func shmAt(id uintptr) uintptr {
	addr, _, _ := unix.Syscall(unix.SYS_SHMAT, id, 0, 0)
	return addr
}

type MemPool struct {
	vertexArrayCnt                   int
	vertexValId, vertexValAddr       []uintptr
	inDId, inDAddr, outDId, outDAddr uintptr
	vertexVal                        [][]uint32
	inD, outD                        []uint32
}

func NewSharedMemPool(requestId string, totalNode int32, vertexArrayCnt int) *MemPool {
	id, _ := strconv.Atoi(requestId)
	vertexValId := make([]uintptr, 0)
	vertexValAddr := make([]uintptr, 0)
	for i := 0; i < vertexArrayCnt; i++ {
		vertexValId = append(vertexValId, shmGet(uintptr(id*10+i), uintptr(totalNode*4)))
		vertexValAddr = append(vertexValAddr, shmAt(vertexValId[i]))
	}
	inDId := shmGet(uintptr(id*10+vertexArrayCnt), uintptr(totalNode*4))
	inDAddr := shmAt(inDId)
	outDId := shmGet(uintptr(id*10+vertexArrayCnt+1), uintptr(totalNode*4))
	outDAddr := shmAt(outDId)
	return &MemPool{
		vertexArrayCnt: vertexArrayCnt,
		vertexValId:    vertexValId,
		vertexValAddr:  vertexValAddr,
		inDId:          inDId,
		inDAddr:        inDAddr,
		outDId:         outDId,
		outDAddr:       outDAddr,
	}
}

func DetachSharedMemPool(m *MemPool) {
	for _, vertexValAddr := range m.vertexValAddr {
		unix.Syscall(unix.SYS_SHMDT, vertexValAddr, 0, 0)
	}
	unix.Syscall(unix.SYS_SHMDT, m.inDAddr, 0, 0)
	unix.Syscall(unix.SYS_SHMDT, m.outDAddr, 0, 0)
}

func NewMemPool(requestId string, totalNode int32, vertexArrayCnt int) *MemPool {
	vertexVal := make([][]uint32, vertexArrayCnt)
	for i := 0; i < vertexArrayCnt; i++ {
		vertexVal[i] = make([]uint32, totalNode)
	}
	m := &MemPool{
		vertexVal: vertexVal,
		inD:       make([]uint32, totalNode),
		outD:      make([]uint32, totalNode),
	}
	m.vertexValAddr = make([]uintptr, 0)
	for i := 0; i < vertexArrayCnt; i++ {
		m.vertexValAddr = append(m.vertexValAddr, uintptr(unsafe.Pointer(&m.vertexVal[i][0])))
	}
	m.inDAddr = uintptr(unsafe.Pointer(&m.inD[0]))
	m.outDAddr = uintptr(unsafe.Pointer(&m.outD[0]))
	return m
}

func (s *MemPool) get(vertex uint32, addr uintptr) uint32 {
	return atomic.LoadUint32((*uint32)(unsafe.Pointer(addr + uintptr(vertex*4))))
}

func (s *MemPool) set(vertex uint32, val uint32, addr uintptr) {
	atomic.StoreUint32((*uint32)(unsafe.Pointer(addr+uintptr(vertex*4))), val)
}

func (s *MemPool) minCas(vertex uint32, val uint32, addr uintptr) bool {
	vertex_addr := (*uint32)(unsafe.Pointer(addr + uintptr(vertex*4)))
	for {
		old := atomic.LoadUint32(vertex_addr)
		if old <= val {
			return false
		}
		if atomic.CompareAndSwapUint32(vertex_addr, old, val) {
			break
		}
	}
	return true
}

func (s *MemPool) floatAdd(vertex uint32, val uint32, addr uintptr) bool {
	vertex_addr := (*uint32)(unsafe.Pointer(addr + uintptr(vertex*4)))
	for {
		old := atomic.LoadUint32(vertex_addr)
		new := math.Float32bits(math.Float32frombits(old) + math.Float32frombits(val))
		if atomic.CompareAndSwapUint32(vertex_addr, old, new) {
			break
		}
	}
	return true
}
