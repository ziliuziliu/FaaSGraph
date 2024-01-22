package graphcalc

import (
	"io"
	"lambda_executor/src/common"
	"lambda_executor/src/graphutil"
	"sync"
)

var VERTEX_ARRAY_TAG_NEW = 0

func Load(u int32, inDegree int32, outDegree int32, graph *graphutil.Graph, firstActivate *common.Bitmap, param map[string]any) {
	graph.SetVertexVal(0, u, uint32(u))
	firstActivate.Add(int(u))
}

func Calc(activeVertex []int32, graph *graphutil.Graph, nextActivate *common.Bitmap, writers []*io.PipeWriter) {
	wg := sync.WaitGroup{}
	calcBlock := 65536
	calcFunc := func(activeVertex []int32) {
		data := make([]uint32, 0, calcBlock*2+1)
		data = append(data, 0)
		for _, u := range activeVertex {
			val := graph.GetVertexVal(0, u)
			newVal := val
			inEdges := graph.InEdge(u)
			for i := 0; i < len(inEdges); i++ {
				v := inEdges[i]
				if vval := graph.GetVertexVal(0, v); vval < newVal {
					newVal = vval
				}
			}
			if newVal < val {
				graph.SetVertexVal(0, u, newVal)
				nextActivate.Add(int(u))
				data = append(data, uint32(u), newVal)
			}
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
	for i := 0; i < len(activeVertex); i += calcBlock {
		wg.Add(1)
		if i+calcBlock < len(activeVertex) {
			go calcFunc(activeVertex[i : i+calcBlock])
		} else {
			go calcFunc(activeVertex[i:])
		}
	}
	wg.Wait()
}
