package graphcalc

import (
	"io"
	"lambda_executor/src/common"
	"lambda_executor/src/graphutil"
	"math"
	"sync"
)

var inf = uint32(math.MaxInt32) - 1000
var VERTEX_ARRAY_TAG_NEW = 0

func Load(u int32, inDegree int32, outDegree int32, graph *graphutil.Graph, firstActivate *common.Bitmap, param map[string]any) {
	if u != 0 {
		graph.SetVertexVal(0, u, inf)
	} else {
		graph.SetVertexVal(0, u, 0)
		firstActivate.Add(0)
	}
}

func Calc(activeVertex []int32, graph *graphutil.Graph, nextActivate *common.Bitmap, writers []*io.PipeWriter) {
	wg := sync.WaitGroup{}
	sendFunc := func(data []uint32) {
		if len(data) > 1 && len(writers) > 0 {
			data[0] = uint32((len(data) - 1) / 2)
			byteData := common.Uint32Slice2ByteSlice(data)
			for _, writer := range writers {
				writer.Write(byteData)
			}
		}
		wg.Done()
	}
	data := make([]uint32, 0, len(activeVertex)*2+1)
	data = append(data, 0)
	startNode := graph.GetStartNode()
	size := graph.GetSize()
	sendOff := 0
	sendThreshold := 65536
	for len(activeVertex) > 0 {
		x := (activeVertex)[0]
		activeVertex = activeVertex[1:]
		val := graph.GetVertexVal(0, x)
		outEdges := graph.OutEdge(x)
		for i := 0; i < len(outEdges); i += 2 {
			v, w := outEdges[i], outEdges[i+1]
			if graph.GetVertexVal(0, v) > val+uint32(w) {
				graph.AggregateVertexVal(0, uint32(v), val+uint32(w))
				if v >= startNode && v < startNode+size {
					activeVertex = append(activeVertex, v)
				} else {
					nextActivate.Add(int(v))
				}
				data = append(data, uint32(v), val+uint32(w))
				if len(data) >= sendOff+sendThreshold {
					wg.Add(1)
					go sendFunc(data[sendOff:])
					sendOff = len(data)
					data = append(data, 0)
				}
			}
		}
	}
	wg.Add(1)
	go sendFunc(data[sendOff:])
	wg.Wait()
}
