package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

var txtPath string
var delimiter = " "
var partition int
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

func readOriginalGraph() ([]int32, []int32) {
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
		inDegree[v]++
		outDegree[u]++
	}
	originalF.Close()
	return inDegree, outDegree
}

func split(inDegree, outDegree []int32) []int32 {
	limit := totalEdge * 2 / partition
	startNodes := make([]int32, 0, partition+1)
	for start := 0; start < totalNode; {
		startNodes = append(startNodes, int32(start))
		cur := 0
		for ; start < totalNode; start++ {
			cur += int(inDegree[start] + outDegree[start])
			if cur > limit {
				start++
				break
			}
		}
	}
	startNodes = append(startNodes, int32(totalNode))
	return startNodes
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

func buildSplit(startNodes []int32) {
	fmt.Println("======Write StartNode======")
	startNodeF, err := os.Create(filepath.Join(dataPath, "start_nodes_"+strconv.Itoa(partition)+".bin"))
	if err != nil {
		panic(err)
	}
	startNodeF.Write(Int32Slice2ByteSlice(startNodes))
	startNodeF.Close()
}

func main() {
	txtPath = os.Args[1]
	dataPath = os.Args[2]
	if os.Args[3] == "WEIGHTED" {
		weighted = true
	} else {
		weighted = false
	}
	totalNode, _ = strconv.Atoi(os.Args[4])
	partition, _ = strconv.Atoi(os.Args[5])
	inDegree, outDegree := readOriginalGraph()
	startNodes := split(inDegree, outDegree)
	buildSplit(startNodes)
}
