package main

import (
	"fmt"
	"math"
	"strconv"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	IPC_CREATE = 00001000
	IPC_RMID   = 0
)

type MemPool struct {
	vertexValId   []uintptr
	inDId         uintptr
	outDId        uintptr
	vertexValAddr []uintptr
	inDAddr       uintptr
	outDAddr      uintptr
}

func shmGet(key uintptr, length uintptr) uintptr {
	id, _, err := unix.Syscall(unix.SYS_SHMGET, key, length, IPC_CREATE|0666)
	if err != 0 {
		id, _, _ = unix.Syscall(unix.SYS_SHMGET, key, 1, IPC_CREATE|0666)
		unix.Syscall(unix.SYS_SHMCTL, id, IPC_RMID, 0)
		id, _, _ = unix.Syscall(unix.SYS_SHMGET, key, length, IPC_CREATE|0666)
	}
	return id
}

func shmAt(id uintptr) uintptr {
	addr, _, _ := unix.Syscall(unix.SYS_SHMAT, id, 0, 0)
	return addr
}

func NewMemPool(requestId string, totalNode int, vertexArrayCnt int) *MemPool {
	fmt.Println("allocating memory for", requestId, "length", totalNode)
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
		vertexValId:   vertexValId,
		inDId:         inDId,
		outDId:        outDId,
		vertexValAddr: vertexValAddr,
		inDAddr:       inDAddr,
		outDAddr:      outDAddr,
	}

}

func (m *MemPool) Set(vertex uint32, val uint32, addr uintptr) {
	atomic.StoreUint32((*uint32)(unsafe.Pointer(addr+uintptr(vertex*4))), val)
}

func (m *MemPool) MinCas(vertex uint32, val uint32, addr uintptr) {
	vertex_addr := (*uint32)(unsafe.Pointer(addr + uintptr(vertex*4)))
	for {
		old := atomic.LoadUint32(vertex_addr)
		if old <= val {
			break
		}
		if atomic.CompareAndSwapUint32(vertex_addr, old, val) {
			break
		}
	}
}

func (m *MemPool) FloatAdd(vertex uint32, val uint32, addr uintptr) bool {
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

func (m *MemPool) AggregateVertexVal(tag int, aggregateOp string, data []uint32) {
	if aggregateOp == "MIN_CAS" {
		for i := 0; i < len(data); i += 2 {
			m.MinCas(data[i], data[i+1], m.vertexValAddr[tag])
		}
	} else if aggregateOp == "FLOAT_ADD" {
		for i := 0; i < len(data); i += 2 {
			m.FloatAdd(data[i], data[i+1], m.vertexValAddr[tag])
		}
	} else {
		for i := 0; i < len(data); i += 2 {
			m.Set(data[i], data[i+1], m.vertexValAddr[tag])
		}
	}
}

func (m *MemPool) ClearMem() {
	for _, vertexValAddr := range m.vertexValAddr {
		unix.Syscall(unix.SYS_SHMDT, vertexValAddr, 0, 0)
	}
	for _, vertexValId := range m.vertexValId {
		_, _, err := unix.Syscall(unix.SYS_SHMCTL, vertexValId, 0, 0)
		if err != 0 {
			panic(err)
		}
	}
	unix.Syscall(unix.SYS_SHMDT, m.inDAddr, 0, 0)
	_, _, err := unix.Syscall(unix.SYS_SHMCTL, m.inDId, 0, 0)
	if err != 0 {
		panic(err)
	}
	unix.Syscall(unix.SYS_SHMDT, m.outDAddr, 0, 0)
	_, _, err = unix.Syscall(unix.SYS_SHMCTL, m.outDId, 0, 0)
	if err != 0 {
		panic(err)
	}
}
