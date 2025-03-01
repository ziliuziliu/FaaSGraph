package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

var txtPath string
var delimiter = " "
var weighted bool
var dataPath string
var directed = true
var totalNode int
var totalEdge = 0

func addEdge(u, v int32, inEdgeList, outEdgeList [][]int32) {
	inEdgeList[v] = append(inEdgeList[v], u)
	if weighted {
		inEdgeList[v] = append(inEdgeList[v], (u+v)%100)
	}
	outEdgeList[u] = append(outEdgeList[u], v)
	if weighted {
		outEdgeList[u] = append(outEdgeList[u], (u+v)%100)
	}
}

func readOriginalGraph() ([][]int32, [][]int32, []int32, []int32) {
	// degree
	fmt.Println("======Calculating Degree======")
	inDegree := make([]int32, totalNode)
	outDegree := make([]int32, totalNode)
	originalF, err := os.Open(txtPath)
	if err != nil {
		panic(err)
	}
	r := bufio.NewScanner(originalF)
	for r.Scan() {
		nodes := strings.Split(r.Text(), delimiter)
		u, _ := strconv.Atoi(nodes[0])
		v, _ := strconv.Atoi(nodes[1])
		totalEdge++
		if totalEdge <= 10 {
			fmt.Println(u, v)
		}
		if totalEdge%10000000 == 0 {
			fmt.Println("current", totalEdge, "edges")
		}
		if !directed {
			inDegree[u]++
			outDegree[v]++
			totalEdge++
		}
		inDegree[v]++
		outDegree[u]++
	}
	originalF.Close()

	// csr
	fmt.Println("======Deriving CSR======")
	totalEdge = 0
	inEdgeList := make([][]int32, totalNode)
	outEdgeList := make([][]int32, totalNode)
	for i := 0; i < totalNode; i++ {
		inEdgeList[i] = make([]int32, 0, inDegree[i])
	}
	for i := 0; i < totalNode; i++ {
		outEdgeList[i] = make([]int32, 0, outDegree[i])
	}
	originalF, err = os.Open(txtPath)
	if err != nil {
		panic(err)
	}
	r = bufio.NewScanner(originalF)
	for r.Scan() {
		nodes := strings.Split(r.Text(), delimiter)
		u, _ := strconv.Atoi(nodes[0])
		v, _ := strconv.Atoi(nodes[1])
		totalEdge++
		if totalEdge <= 10 {
			fmt.Println(u, v)
		}
		if totalEdge%10000000 == 0 {
			fmt.Println("current", totalEdge, "edges")
		}
		addEdge(int32(u), int32(v), inEdgeList, outEdgeList)
		if !directed {
			totalEdge++
			addEdge(int32(v), int32(u), inEdgeList, outEdgeList)
		}
	}
	originalF.Close()
	runtime.GC()
	return inEdgeList, outEdgeList, inDegree, outDegree
}

func Int64Slice2ByteSlice(a []int64) []byte {
	var b []byte
	int64Slice := reflect.ValueOf(a)
	byteSlice := reflect.New(reflect.TypeOf(b))
	header := (*reflect.SliceHeader)(unsafe.Pointer(byteSlice.Pointer()))
	header.Cap = int64Slice.Cap() * 8
	header.Len = int64Slice.Len() * 8
	header.Data = uintptr(int64Slice.Pointer())
	return byteSlice.Elem().Bytes()
}

func Int32Slice2ByteSlice(a []int32) []byte {
	var b []byte
	int32Slice := reflect.ValueOf(a)
	byteSlice := reflect.New(reflect.TypeOf(b))
	header := (*reflect.SliceHeader)(unsafe.Pointer(byteSlice.Pointer()))
	header.Cap = int32Slice.Cap() * 4
	header.Len = int32Slice.Len() * 4
	header.Data = uintptr(int32Slice.Pointer())
	return byteSlice.Elem().Bytes()
}

func buildGraph(inEdgeList [][]int32, outEdgeList [][]int32, inDegree []int32, outDegree []int32) {
	err := os.RemoveAll(dataPath)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(dataPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	fmt.Println("======Write InOffset======")
	inOffset := make([]int64, len(inDegree)+1)
	for i, degree := range inDegree {
		inOffset[i+1] = inOffset[i] + int64(degree)
	}
	inOffsetF, err := os.Create(filepath.Join(dataPath, "in_offset.bin"))
	if err != nil {
		panic(err)
	}
	inOffsetF.Write(Int64Slice2ByteSlice(inOffset))
	inOffsetF.Close()
	fmt.Println(inOffset[totalNode])
	fmt.Println("======Write OutOffset======")
	outOffset := make([]int64, len(outDegree)+1)
	for i, degree := range outDegree {
		outOffset[i+1] = outOffset[i] + int64(degree)
	}
	outOffsetF, err := os.Create(filepath.Join(dataPath, "out_offset.bin"))
	if err != nil {
		panic(err)
	}
	outOffsetF.Write(Int64Slice2ByteSlice(outOffset))
	outOffsetF.Close()
	fmt.Println(outOffset[totalNode])
	fmt.Println("======Write InEdge======")
	inEdgeF, err := os.Create(filepath.Join(dataPath, "in_edge.bin"))
	if err != nil {
		panic(err)
	}
	for _, edgeList := range inEdgeList {
		if len(edgeList) > 0 {
			inEdgeF.Write(Int32Slice2ByteSlice(edgeList))
		}
	}
	inEdgeF.Close()
	fmt.Println("======Write OutEdge======")
	outEdgeF, err := os.Create(filepath.Join(dataPath, "out_edge.bin"))
	if err != nil {
		panic(err)
	}
	for _, edgeList := range outEdgeList {
		if len(edgeList) > 0 {
			outEdgeF.Write(Int32Slice2ByteSlice(edgeList))
		}
	}
	outEdgeF.Close()
}

func main() {
	txtPath = os.Args[1]
	dataPath = os.Args[2]
	if os.Args[3] == "WEIGHTED" {
		weighted = true
	} else {
		weighted = false
	}
	if os.Args[4] == "DIRECTED" {
		directed = true
	} else {
		directed = false
	}
	totalNode, _ = strconv.Atoi(os.Args[5])
	inEdgeList, outEdgeList, inDegree, outDegree := readOriginalGraph()
	buildGraph(inEdgeList, outEdgeList, inDegree, outDegree)
}
