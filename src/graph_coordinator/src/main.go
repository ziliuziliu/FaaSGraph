package main

import (
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

func Status(writer http.ResponseWriter, request *http.Request) {
	writer.Write(OKResult())
}

func Init(writer http.ResponseWriter, request *http.Request) {
	writer.Write(OKResult())
}

func Run(writer http.ResponseWriter, request *http.Request) {
	req, _ := ioutil.ReadAll(request.Body)
	containerSet := ContainerSet{}
	err := json.Unmarshal(req, &containerSet)
	if err != nil {
		panic(err)
	}
	preprocessTime, computeTime, flushTime, comm, maxComm, minComm := containerSet.run()
	result := map[string]float64{
		"preprocess": preprocessTime,
		"query":      computeTime,
		"store":      flushTime,
		"comm":       float64(comm) * 8 / 1024 / 1024 / 1024,
		"maxc":       float64(maxComm) * 8 / 1024 / 1024 / 1024,
		"minc":       float64(minComm) * 8 / 1024 / 1024 / 1024,
	}
	resultStr, _ := json.Marshal(result)
	writer.Write(resultStr)
}

func main() {
	http.HandleFunc("/status", Status)
	http.HandleFunc("/init", Init)
	http.HandleFunc("/run", Run)
	fmt.Println("server started at 0.0.0.0:5000")
	http.ListenAndServe(":5000", nil)
}
