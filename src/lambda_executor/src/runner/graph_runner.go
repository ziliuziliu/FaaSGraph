package runner

import (
	"fmt"
	"io"
	"lambda_executor/src/common"
	"lambda_executor/src/graphcalc"
	"lambda_executor/src/graphutil"
	"net/http"
	"strconv"
	"sync"
)

var mutex sync.Mutex = sync.Mutex{}

type GraphRunner struct {
	graph             string
	ip                string
	port              int32
	app               string
	function          string
	totalNode         int32
	config            map[string]any
	graphConfig       map[string]any
	graphs            map[int32]*graphutil.Graph
	workerIPList      []string
	containerAddrList []string
	startNodeList     []int32
	shareMemory       bool
}

func NewGraphRunner(graph string, ip string, port int32, app string, function string, config map[string]any, workerIPList []string, containerAddrList []string, startNodeList []int32) *GraphRunner {
	r := &GraphRunner{
		graph:             graph,
		ip:                ip,
		port:              port,
		app:               app,
		function:          function,
		config:            config,
		graphs:            make(map[int32]*graphutil.Graph),
		workerIPList:      workerIPList,
		containerAddrList: containerAddrList,
		startNodeList:     startNodeList,
		shareMemory:       config["SHARE_MEMORY"].(bool),
	}
	fmt.Println(containerAddrList, startNodeList)
	r.graphConfig = r.config["GRAPH"].(map[string]any)[graph].(map[string]any)
	r.totalNode = int32(r.graphConfig["TOTAL_NODE"].(float64))
	var err error
	if err != nil {
		panic(err)
	}
	return r
}

func (r *GraphRunner) pushVertexVals(wg *sync.WaitGroup, reader io.Reader, requestId string, remoteAddr string, startNode int32, graph *graphutil.Graph) {
	pushUrl := fmt.Sprintf("http://%s/set_vertex_info", remoteAddr)
	req, _ := http.NewRequest("POST", pushUrl, reader)
	req.Header.Add("Request-Id", requestId)
	req.Header.Add("Start-Node", strconv.Itoa(int(startNode)))
	req.Header.Add("Aggregate-Op", graph.GetAggregateOp())
	req.Header.Add("Tag", strconv.Itoa(graphcalc.VERTEX_ARRAY_TAG_NEW))
	rep, _ := http.DefaultClient.Do(req)
	io.Copy(io.Discard, rep.Body)
	rep.Body.Close()
	wg.Done()
}

func (r *GraphRunner) createWriters(wg *sync.WaitGroup, requestId string, graph *graphutil.Graph) []*io.PipeWriter {
	var writers []*io.PipeWriter
	if r.shareMemory {
		writers = make([]*io.PipeWriter, 0, len(r.workerIPList)-1)
		for _, workerIP := range r.workerIPList {
			if workerIP != r.ip {
				reader, writer := io.Pipe()
				writers = append(writers, writer)
				wg.Add(1)
				go r.pushVertexVals(wg, reader, requestId, workerIP+":20000", 0, graph)
			}
		}
	} else {
		writers = make([]*io.PipeWriter, 0, len(r.containerAddrList)-1)
		for i, containerAddr := range r.containerAddrList {
			if r.startNodeList[i] != graph.GetStartNode() {
				reader, writer := io.Pipe()
				writers = append(writers, writer)
				wg.Add(1)
				go r.pushVertexVals(wg, reader, requestId, containerAddr, r.startNodeList[i], graph)
			}
		}
	}
	return writers
}

func (r *GraphRunner) LoadKey(key string, size int32, inEdge bool, outEdge bool, aggregateOp string) {
	startNode, _ := strconv.Atoi(key)
	graph := graphutil.NewGraph(r.graph, r.ip, r.port, r.app, r.function, int32(startNode), size, r.totalNode, inEdge, outEdge, aggregateOp, r.config, r.graphConfig)
	graph.LoadGraph()
	mutex.Lock()
	r.graphs[int32(startNode)] = graph
	mutex.Unlock()
}

func (r *GraphRunner) LoadMemPool(requestId string, key string, vertexArrayCnt int) {
	startNodeN, _ := strconv.Atoi(key)
	startNode := int32(startNodeN)
	graph := r.graphs[startNode]
	graph.LoadMemPool(requestId, vertexArrayCnt)
}

func (r *GraphRunner) LoadMem(requestId string, key string, param map[string]any) []uint32 {
	startNodeN, _ := strconv.Atoi(key)
	startNode := int32(startNodeN)
	graph := r.graphs[startNode]
	block := int32(r.config["CALC_BLOCK"].(float64))
	size := graph.GetSize()
	activate := common.NewBitmap(int(r.totalNode))
	wg := &sync.WaitGroup{}
	wg2 := &sync.WaitGroup{}
	writers := r.createWriters(wg2, requestId, graph)
	loadMemFunc := func(vertex []int32) {
		data := make([]uint32, 0, len(vertex)*2+1)
		data = append(data, uint32(len(vertex)))
		for _, v := range vertex {
			graphcalc.Load(v, graph.CalcInDegree(v), graph.CalcOutDegree(v), graph, activate, param)
			data = append(data, uint32(v), graph.GetVertexVal(0, v))
		}
		for _, writer := range writers {
			writer.Write(common.Uint32Slice2ByteSlice(data))
		}
		wg.Done()
	}
	for i := int32(0); i < size; i += block {
		wg.Add(1)
		vertex := make([]int32, 0, block)
		for j := int32(0); i+j < size && j < block; j++ {
			vertex = append(vertex, startNode+i+j)
		}
		go loadMemFunc(vertex)
	}
	wg.Wait()
	for _, writer := range writers {
		writer.Close()
	}
	wg2.Wait()
	return activate.Data
}

func (r *GraphRunner) Run(requestId string, startNode int32, activeVertex []byte, noChange bool) ([]byte, map[string]float64) {
	active := common.NewBitmapWith(common.ByteSlice2Uint32Slice(activeVertex))
	graph := r.graphs[startNode]
	graph.ResetPerformanceMetrics()
	start := common.GetTime()
	nextActivate, gasRun := r.gas(requestId, active, graph, noChange)
	graph.SetPerformanceMetrics("run_overall", common.GetTime()-start)
	graph.SetPerformanceMetrics("gas_run", gasRun)
	return common.Uint32Slice2ByteSlice(nextActivate.Data), graph.GetPerformanceMetrics()
}

func (r *GraphRunner) SetVertexInfo(requestId string, startNode int32, tag int, aggregateOp string, data []uint32) {
	graph, ok := r.graphs[startNode]
	if !ok {
		panic("wrong graph partition")
	}
	if aggregateOp == "MIN_CAS" {
		for i := 0; i < len(data); i += 2 {
			graph.AggregateVertexVal(tag, data[i], data[i+1])
		}
	} else {
		for i := 0; i < len(data); i += 2 {
			graph.SetVertexVal(tag, int32(data[i]), data[i+1])
		}
	}
}

func (r *GraphRunner) gas(requestId string, activeVertex *common.Bitmap, graph *graphutil.Graph, noChange bool) (*common.Bitmap, float64) {
	nextActivate := common.NewBitmap(int(r.totalNode))
	block := int(r.config["CALC_BLOCK"].(float64))
	l := graph.GetSize()
	wg := &sync.WaitGroup{}
	writers := r.createWriters(wg, requestId, graph)
	active := make([]int32, 0, block)
	startNode := graph.GetStartNode()
	for v := startNode; v < startNode+l; v++ {
		if activeVertex.Get(int(v)) {
			active = append(active, v)
		}
	}
	var gasRun float64
	if len(active) != 0 {
		start := common.GetTime()
		graphcalc.Calc(active, graph, nextActivate, writers)
		gasRun = common.GetTime() - start
	}
	for _, writer := range writers {
		if writer != nil {
			writer.Close()
		}
	}
	wg.Wait()
	return nextActivate, gasRun
}

func (r *GraphRunner) Flush() map[string]any {
	wg := sync.WaitGroup{}
	saveFunc := func(graph *graphutil.Graph) {
		// graph.Save() // TODO: when high load, no flush, just for experiments
		graph.DetachMemPool()
		wg.Done()
	}
	for _, g := range r.graphs {
		wg.Add(1)
		go saveFunc(g)
	}
	wg.Wait()
	return map[string]any{"status": "OK"}
}
