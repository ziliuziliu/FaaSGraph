package graphutil

import (
	"fmt"
	"lambda_executor/src/common"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Graph struct {
	graph            string
	ip               string
	port             int32
	app              string
	function         string
	dataDir          string
	startNode        int32
	size             int32
	totalNode        int32
	inEdge           bool
	inCSR            []int32
	inOffset         []int32
	outEdge          bool
	outCSR           []int32
	outOffset        []int32
	memPool          *MemPool
	config           map[string]any
	graphConfig      map[string]any
	weighted         bool
	aggregateOp      string
	aggregateFunc    func(uint32, uint32, uintptr) bool
	partitionPerHost int
	shareMemory      bool
	performance      map[string]float64
}

func NewGraph(graph string, ip string, port int32, app string, function string, startNode int32, size int32, totalNode int32, inEdge bool, outEdge bool, aggregateOp string, config map[string]any, graphConfig map[string]any) *Graph {
	return &Graph{
		graph:            graph,
		ip:               ip,
		port:             port,
		app:              app,
		function:         function,
		dataDir:          filepath.Join("/lambda_executor/data", strconv.Itoa(int(startNode))),
		startNode:        startNode,
		size:             size,
		totalNode:        totalNode,
		inEdge:           inEdge,
		outEdge:          outEdge,
		aggregateOp:      aggregateOp,
		config:           config,
		graphConfig:      graphConfig,
		weighted:         graphConfig["WEIGHTED"].(bool),
		partitionPerHost: int(config["MAX_SLOT_PER_NODE"].(float64)),
		shareMemory:      config["SHARE_MEMORY"].(bool),
		performance:      make(map[string]float64),
	}
}

// func (g *Graph) readFromFile(filename string) []byte {
// 	f, err := directio.OpenFile(filename, os.O_RDONLY, 0666)
// 	if err != nil {
// 		panic(err)
// 	}
// 	stat, _ := f.Stat()
// 	size := int(stat.Size())
// 	alignedSize := size
// 	if alignedSize%directio.AlignSize != 0 {
// 		alignedSize = (alignedSize/directio.AlignSize + 1) * directio.AlignSize
// 	}
// 	b := directio.AlignedBlock(alignedSize)
// 	buf, cur := b, 0
// 	for {
// 		n, err := f.Read(buf)
// 		if err != nil {
// 			panic(err)
// 		}
// 		cur += n
// 		if cur == size {
// 			break
// 		}
// 		buf = buf[n:]
// 	}
// 	b = b[:size]
// 	return b
// }

func (g *Graph) readFromFile(filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	stat, _ := f.Stat()
	size := int(stat.Size())
	b := make([]byte, size)
	buf, cur := b, 0
	for {
		n, _ := f.Read(buf)
		cur += n
		if cur == int(size) {
			break
		}
		buf = buf[n:]
	}
	b = b[:size]
	return b
}

func (g *Graph) readSegmentFromFile(filename string, offset int64, len int64) []byte {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	f.Seek(offset, 0)
	b := make([]byte, len)
	buf, cur := b, 0
	for {
		n, _ := f.Read(buf)
		cur += n
		if cur == int(len) {
			break
		}
		buf = buf[n:]
	}
	b = b[:len]
	return b
}

func (g *Graph) LoadGraph() {
	el := int64(1)
	if g.weighted {
		el = 2
	}
	dirPath := "/lambda_executor/data"
	startNodeList := common.ByteSlice2Int32Slice(
		g.readFromFile(
			filepath.Join(dirPath, "start_nodes_"+strconv.Itoa(int(g.graphConfig["PARTITION"].(float64)))+".bin"),
		),
	)
	startNodeList = append(startNodeList, g.totalNode)
	startNodeIndex := -1
	for i, startNode := range startNodeList {
		if g.startNode == startNode {
			startNodeIndex = i
			break
		}
	}
	loadOffsetFunc := func(filename string) []byte {
		return g.readSegmentFromFile(
			filename,
			int64(g.startNode)*8,
			int64(startNodeList[startNodeIndex+1]-g.startNode+1)*8,
		)
	}
	inOffset := common.ByteSlice2Int64Slice(loadOffsetFunc(filepath.Join(dirPath, "in_offset.bin")))
	outOffset := common.ByteSlice2Int64Slice(loadOffsetFunc(filepath.Join(dirPath, "out_offset.bin")))
	wg := sync.WaitGroup{}
	loadCSRFunc := func(filename string, offset int64, len int64, CSR *[]int32) {
		*CSR = common.ByteSlice2Int32Slice(g.readSegmentFromFile(filename, offset, len))
		wg.Done()
	}
	if g.inEdge {
		wg.Add(1)
		go loadCSRFunc(
			filepath.Join(dirPath, "in_edge.bin"),
			inOffset[0]*4*el,
			(inOffset[len(inOffset)-1]-inOffset[0])*4*el,
			&g.inCSR,
		)
	}
	if g.outEdge {
		wg.Add(1)
		go loadCSRFunc(
			filepath.Join(dirPath, "out_edge.bin"),
			outOffset[0]*4*el,
			(outOffset[len(outOffset)-1]-outOffset[0])*4*el,
			&g.outCSR,
		)
	}
	g.inOffset = make([]int32, len(inOffset))
	for i, offset := range inOffset {
		g.inOffset[i] = int32((offset - inOffset[0]) * el)
	}
	g.outOffset = make([]int32, len(outOffset))
	for i, offset := range outOffset {
		g.outOffset[i] = int32((offset - outOffset[0]) * el)
	}
	wg.Wait()
}

// func (g *Graph) LoadGraph() {
// 	g.el = 2
// 	if g.weighted {
// 		g.el = 3
// 	}
// 	dirPath := fmt.Sprintf("/lambda_executor/data/%d", g.startNode)
// 	wg := sync.WaitGroup{}
// 	loadFunc := func(mode int, file string, offset []int32, CSR *[]int32) {
// 		edgeBuffer := common.ByteSlice2Int32Slice(g.readFromFile(filepath.Join(dirPath, file)))
// 		*CSR = make([]int32, int(offset[len(offset)-1]))
// 		pos := make([]int32, g.size+1)
// 		copy(pos, offset)
// 		for i := 0; i < len(edgeBuffer); i += g.el {
// 			u, v := edgeBuffer[i], edgeBuffer[i+1]
// 			(*CSR)[pos[u-g.startNode]] = v
// 			pos[u-g.startNode]++
// 			if g.weighted {
// 				(*CSR)[pos[u-g.startNode]] = edgeBuffer[i+2]
// 				pos[u-g.startNode]++
// 			}
// 		}
// 		wg.Done()
// 	}
// 	loadOffsetFunc := func(file string) []int32 {
// 		offset := common.ByteSlice2Int32Slice(g.readFromFile(filepath.Join(dirPath, file)))
// 		if g.weighted {
// 			for i := 0; i < int(g.size); i++ {
// 				offset[i] *= 2
// 			}
// 		}
// 		for i := 1; i < len(offset); i++ {
// 			offset[i] += offset[i-1]
// 		}
// 		offset = append(offset, 0)
// 		for i := len(offset) - 1; i >= 1; i-- {
// 			offset[i] = offset[i-1]
// 		}
// 		offset[0] = 0
// 		return offset
// 	}
// 	loadStartNodeListFunc := func(file string) []int32 {
// 		return common.ByteSlice2Int32Slice(g.readFromFile(file))
// 	}
// 	g.inOffset = loadOffsetFunc("in_degree.bin")
// 	g.outOffset = loadOffsetFunc("out_degree.bin")
// 	g.startNodeList = loadStartNodeListFunc("/lambda_executor/data/start_nodes.bin")
// 	if g.inEdge {
// 		wg.Add(1)
// 		go loadFunc(1, "in.bin", g.inOffset, &g.inCSR)
// 	}
// 	if g.outEdge {
// 		wg.Add(1)
// 		go loadFunc(2, "out.bin", g.outOffset, &g.outCSR)
// 	}
// 	wg.Wait()
// 	runtime.GC()
// }

func (g *Graph) LoadMemPool(requestId string, vertexValCnt int) {
	if g.shareMemory {
		g.memPool = NewSharedMemPool(requestId, g.totalNode, vertexValCnt)
	} else {
		g.memPool = NewMemPool(requestId, g.totalNode, vertexValCnt)
	}
	if g.aggregateOp == "MIN_CAS" {
		g.aggregateFunc = g.memPool.minCas
	} else if g.aggregateOp == "FLOAT_ADD" {
		g.aggregateFunc = g.memPool.floatAdd
	}
}

func (g *Graph) DetachMemPool() {
	if g.shareMemory {
		DetachSharedMemPool(g.memPool)
	}
}

func (g *Graph) ResetPerformanceMetrics() {
	for k := range g.performance {
		delete(g.performance, k)
	}
}

func (g *Graph) GetPerformanceMetrics() map[string]float64 {
	return g.performance
}

func (g *Graph) SetPerformanceMetrics(key string, time float64) {
	g.performance[key] = time
}

func (g *Graph) GetStartNode() int32 {
	return g.startNode
}

func (g *Graph) GetSize() int32 {
	return g.size
}

func (g *Graph) GetWeighted() bool {
	return g.weighted
}

func (g *Graph) GetAggregateOp() string {
	return g.aggregateOp
}

func (g *Graph) GetTotalNode() int32 {
	return g.totalNode
}

func (g *Graph) InEdge(node int32) []int32 {
	return g.inCSR[g.inOffset[node-g.startNode]:g.inOffset[node-g.startNode+1]]
}

func (g *Graph) OutEdge(node int32) []int32 {
	return g.outCSR[g.outOffset[node-g.startNode]:g.outOffset[node-g.startNode+1]]
}

func (g *Graph) GetInDegree(node int32) int32 {
	return int32(g.memPool.get(uint32(node), g.memPool.inDAddr))
}

func (g *Graph) SetInDegree(node int32, inDegree int32) {
	g.memPool.set(uint32(node), uint32(inDegree), g.memPool.outDAddr)
}

func (g *Graph) CalcInDegree(node int32) int32 {
	len := g.inOffset[node-g.startNode+1] - g.inOffset[node-g.startNode]
	if g.weighted {
		len /= 2
	}
	return len
}

func (g *Graph) GetOutDegree(node int32) int32 {
	return int32(g.memPool.get(uint32(node), g.memPool.outDAddr))
}

func (g *Graph) SetOutDegree(node int32, outDegree int32) {
	g.memPool.set(uint32(node), uint32(outDegree), g.memPool.outDAddr)
}

func (g *Graph) CalcOutDegree(node int32) int32 {
	len := g.outOffset[node-g.startNode+1] - g.outOffset[node-g.startNode]
	if g.weighted {
		len /= 2
	}
	return len
}

func (g *Graph) GetVertexVal(tag int, node int32) uint32 {
	return g.memPool.get(uint32(node), g.memPool.vertexValAddr[tag])
}

func (g *Graph) AggregateVertexVal(tag int, vertex uint32, val uint32) bool {
	return g.aggregateFunc(vertex, val, g.memPool.vertexValAddr[tag])
}

func (g *Graph) SetVertexVal(tag int, v int32, val uint32) {
	g.memPool.set(uint32(v), val, g.memPool.vertexValAddr[tag])
}

func (g *Graph) Save() {
	resultName := fmt.Sprintf("/lambda_executor/data/%d/vertex_val.bin", g.startNode)
	data := make([]byte, 0, g.size*4)
	b := make([]byte, 4)
	for i := g.startNode; i < g.startNode+g.size; i++ {
		common.Uint32ToBytes(g.GetVertexVal(0, i), b)
		data = append(data, b...)
	}
	f, err := os.OpenFile(resultName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
	f.Close()
}
