package graphcalc

import (
	"io"
	"lambda_executor/src/common"
	"lambda_executor/src/graphutil"
	"sync"
)

var VERTEX_ARRAY_TAG_NEW = 0

func Load(u int32, inDegree int32, outDegree int32, graph *graphutil.Graph, firstActivate *common.Bitmap, param map[string]any) {
	beginner := int32(param["source"].(float64))
	if u != beginner {
		graph.SetVertexVal(0, u, 1000000000)
	} else {
		graph.SetVertexVal(0, u, 0)
		firstActivate.Add(int(beginner))
	}
}

func Calc(activeVertex []int32, graph *graphutil.Graph, nextActivate *common.Bitmap, writers []*io.PipeWriter) {
	data := make([]uint32, 0, len(activeVertex)*2+1)
	data = append(data, 0)
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
	startNode := graph.GetStartNode()
	size := graph.GetSize()
	sendOff := 0
	sendThreshold := 65536
	for len(activeVertex) > 0 {
		x := (activeVertex)[0]
		activeVertex = activeVertex[1:]
		outEdges := graph.OutEdge(x)
		for i := 0; i < len(outEdges); i++ {
			v := outEdges[i]
			if graph.GetVertexVal(0, v) == 1000000000 {
				graph.SetVertexVal(0, v, uint32(x))
				if v >= startNode && v < startNode+size {
					activeVertex = append(activeVertex, v)
				} else {
					nextActivate.Add(int(v))
				}
				data = append(data, uint32(v), uint32(x))
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
