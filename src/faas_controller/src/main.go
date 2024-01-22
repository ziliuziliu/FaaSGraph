package main

import (
	"application/src/application"
	"application/src/common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func OKResult() []byte {
	result := map[string]string{"status": "OK"}
	rep, _ := json.Marshal(result)
	return rep
}

func RunGraph(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	r := common.GraphComputeRequest{}
	err := json.Unmarshal(req, &r)
	if err != nil {
		panic(err)
	}
	fmt.Println("request", r.RequestId)
	r.GraphConfig = common.LoadGraphConfig(r.Graph)
	gc := application.NewGraphCompute(&r)
	newContainerTime, ioTime, preprocessTime, computeTime, flushTime, comm, maxComm, minComm := gc.Execute()
	result := map[string]float64{
		"startup":    newContainerTime,
		"io":         ioTime,
		"preprocess": preprocessTime,
		"query":      computeTime,
		"store":      flushTime,
		"comm":       comm,
		"maxc":       maxComm,
		"minc":       minComm,
	}
	resultStr, _ := json.Marshal(result)
	writer.Write(resultStr)
}

func main() {
	// f, _ := os.Create("abc.pprof")
	// pprof.StartCPUProfile(f)
	// go func() {
	// 	time.Sleep(time.Minute * 5)
	// 	pprof.StopCPUProfile()
	// 	os.Create("Have_Written")
	// }()
	http.HandleFunc("/run_graph", RunGraph)
	fmt.Println("server started at 0.0.0.0:20001")
	err := http.ListenAndServe(":20001", nil)
	if err != nil {
		panic(err)
	}
}
