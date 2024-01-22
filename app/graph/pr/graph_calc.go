package graphcalc

import (
	"io"
	"math"
	"sync"
	"sync/atomic"

	"lambda_executor/src/common"
	"lambda_executor/src/graphutil"
)

var VERTEX_ARRAY_TAG_NEW = 0
var finished int32

func Load(u int32, inDegree int32, outDegree int32, graph *graphutil.Graph, firstActivate *common.Bitmap, param map[string]any) {
	graph.SetInDegree(u, inDegree)
	graph.SetOutDegree(u, outDegree)
	val := float32(1.0)
	if outDegree > 0 {
		val /= float32(outDegree)
		firstActivate.Add(int(u))
	}
	graph.SetVertexVal(0, u, math.Float32bits(val))
	VERTEX_ARRAY_TAG_NEW = 1
}

func Calc(activeVertex []int32, graph *graphutil.Graph, nextActivate *common.Bitmap, writers []*io.PipeWriter) {
	block := 65536
	wg := sync.WaitGroup{}
	calcFunc := func(activeVertex []int32) {
		data := make([]uint32, 0, block*2+1)
		data = append(data, 0)
		for _, u := range activeVertex {
			var accVal float32 = 0.0
			inEdges := graph.InEdge(u)
			for i := 0; i < len(inEdges); i++ {
				v := inEdges[i]
				accVal += math.Float32frombits(graph.GetVertexVal(VERTEX_ARRAY_TAG_NEW^1, v))
			}
			newVal := (0.15 + 0.85*accVal) / float32(graph.GetOutDegree(u))
			newValBits := math.Float32bits(newVal)
			graph.SetVertexVal(VERTEX_ARRAY_TAG_NEW, u, newValBits)
			data = append(data, uint32(u))
			data = append(data, newValBits)
		}
		if len(data) > 1 && len(writers) > 0 {
			data[0] = uint32((len(data) - 1) / 2)
			byteData := common.Uint32Slice2ByteSlice(data)
			for _, writer := range writers {
				writer.Write(byteData)
			}
		}
		wg.Done()
	}
	for i := 0; i < len(activeVertex); i += block {
		wg.Add(1)
		if i+block < len(activeVertex) {
			go calcFunc(activeVertex[i : i+block])
		} else {
			go calcFunc(activeVertex[i:])
		}
	}
	wg.Wait()
	if ret := atomic.AddInt32(&finished, 1); ret == 2 {
		VERTEX_ARRAY_TAG_NEW ^= 1
		finished = int32(0)
	}
}
