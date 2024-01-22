package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"lambda_executor/src/common"
	"lambda_executor/src/runner"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

var worker *runner.GraphRunner
var f *os.File

func OKResult() []byte {
	result := map[string]string{"status": "OK"}
	rep, _ := json.Marshal(result)
	return rep
}

func Status(writer http.ResponseWriter, request *http.Request) {
	writer.Write(OKResult())
}

func Init(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	data := common.InitRequest{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		panic(err)
	}
	worker = runner.NewGraphRunner(data.Graph, data.Ip, data.Port, data.App, data.Function, data.Config, data.WorkerIPList, data.ContainerAddrList, data.StartNodeList)
	writer.Write(OKResult())
}

func LoadKey(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	data := common.LoadKeyRequest{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		panic(err)
	}
	worker.LoadKey(data.Key, data.Size, data.InEdge, data.OutEdge, data.AggregateOp)
	writer.Write(OKResult())
}

func LoadMemPool(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	data := common.LoadMemPoolRequest{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		panic(err)
	}
	worker.LoadMemPool(data.RequestId, data.Key, data.VertexArrayCnt)
	writer.Write(OKResult())
}

func LoadMem(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	data := common.LoadMemRequest{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		panic(err)
	}
	writer.Write(common.Uint32Slice2ByteSlice(worker.LoadMem(data.RequestId, data.Key, data.Param)))
	// if data.Key == "0" {
	// 	ff := rand.Int63()
	// 	f, err = os.Create("/lambda_executor/data/" + data.Key + "/" + strconv.Itoa(int(ff)))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// }
}

func Run(writer http.ResponseWriter, request *http.Request) {
	requestId := request.Header.Get("Request-Id")
	startNode, _ := strconv.Atoi(request.Header.Get("Start-Node"))
	noChange, _ := strconv.ParseBool(request.Header.Get("No-Change"))
	activeVertex, _ := ioutil.ReadAll(request.Body)
	activate, metrics := worker.Run(requestId, int32(startNode), activeVertex, noChange)
	metricsJsonBody, _ := json.Marshal(metrics)
	writer.Header().Add("Metrics", string(metricsJsonBody))
	writer.Write(activate)
	runtime.GC()
}

func SetVertexInfo(writer http.ResponseWriter, request *http.Request) {
	b := make([]byte, 4)
	requestId := request.Header.Get("Request-Id")
	startNode, _ := strconv.Atoi(request.Header.Get("Start-Node"))
	aggregateOp := request.Header.Get("Aggregate-Op")
	tag, _ := strconv.Atoi(request.Header.Get("Tag"))
	for {
		_, err := request.Body.Read(b)
		if err != nil {
			break
		}
		num := binary.LittleEndian.Uint32(b)
		data := make([]byte, num*8)
		for i := uint32(0); i < num; i++ {
			request.Body.Read(data[8*i : 8*i+4])
			request.Body.Read(data[8*i+4 : 8*i+8])
		}
		worker.SetVertexInfo(requestId, int32(startNode), tag, aggregateOp, common.ByteSlice2Uint32Slice(data))
	}
	writer.Write(OKResult())
}

func Flush(writer http.ResponseWriter, request *http.Request) {
	// pprof.StopCPUProfile()
	// f.Close()
	out := worker.Flush()
	rep, _ := json.Marshal(out)
	writer.Write(rep)
	runtime.GC()
}

func main() {
	port := 5000
	if len(os.Args) > 1 {
		port, _ = strconv.Atoi(os.Args[1])
	}
	http.HandleFunc("/status", Status)
	http.HandleFunc("/init", Init)
	http.HandleFunc("/load_key", LoadKey)
	http.HandleFunc("/load_mem_pool", LoadMemPool)
	http.HandleFunc("/load_mem", LoadMem)
	http.HandleFunc("/run", Run)
	http.HandleFunc("/set_vertex_info", SetVertexInfo)
	http.HandleFunc("/flush", Flush)
	fmt.Println("server started at 0.0.0.0:" + strconv.Itoa(port))
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
